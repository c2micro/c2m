package operator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/c2micro/c2mshr/defaults"
	def "github.com/c2micro/c2mshr/defaults"
	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/c2micro/c2msrv/internal/constants"
	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/ent/beacon"
	"github.com/c2micro/c2msrv/internal/ent/blobber"
	"github.com/c2micro/c2msrv/internal/ent/credential"
	"github.com/c2micro/c2msrv/internal/ent/group"
	"github.com/c2micro/c2msrv/internal/ent/listener"
	"github.com/c2micro/c2msrv/internal/ent/message"
	"github.com/c2micro/c2msrv/internal/ent/operator"
	"github.com/c2micro/c2msrv/internal/ent/task"
	"github.com/c2micro/c2msrv/internal/middleware/grpcauth"
	"github.com/c2micro/c2msrv/internal/pools"
	"github.com/c2micro/c2msrv/internal/shared"
	"github.com/c2micro/c2msrv/internal/utils"
	"github.com/c2micro/c2msrv/internal/version"
	errs "github.com/go-faster/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	operatorv1.UnimplementedOperatorServiceServer
	ctx context.Context
	db  *ent.Client
	lg  *zap.Logger
}

// Обработка первоначального хендшейка от оператора.
func (s *server) Hello(req *operatorv1.HelloRequest, stream operatorv1.OperatorService_HelloServer) error {
	ctx := stream.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("Hello").With(zap.String("username", username))

	// валидируем версию клиента
	if strings.Compare(req.GetVersion(), version.Version()) != 0 {
		lg.Warn(shared.ErrorVersionMismatched, zap.String("version", req.GetVersion()))
		return status.Error(codes.InvalidArgument, shared.ErrorVersionMismatched)
	}

	// проверяем, что оператора нет в сессионной мапе
	if pools.Pool.Hello.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected)
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	cookie := utils.RandString(32)

	// сохраняем оператора в сессионную мапу
	pools.Pool.Hello.Add(cookie, username, stream)

	defer func() {
		// удаление из сессионной мапы
		pools.Pool.Hello.Remove(cookie)
		// отключение оператора от всех подписок
		pools.Pool.DisconnectAll(s.ctx, cookie)

		// создание нового сообщения
		if c, err := s.db.Chat.
			Create().
			SetMessage(username + " logged out").
			Save(s.ctx); err != nil {
			lg.Warn(shared.ErrorSaveChatMessage, zap.Error(err))
		} else {
			// нотификация всех подписчиков о создании нового сообщения
			go pools.Pool.Chat.Send(&operatorv1.ChatResponse{
				CreatedAt: timestamppb.New(c.CreatedAt),
				From:      defaults.ChatSrvFrom,
				Message:   c.Message,
			})
		}
	}()

	// отправляем оператору серверный ответ
	if err := stream.Send(&operatorv1.HelloResponse{
		Response: &operatorv1.HelloResponse_Handshake{
			Handshake: &operatorv1.HandshakeResponse{
				Username: username,
				Time:     timestamppb.Now(),
				Cookie: &operatorv1.SessionCookie{
					Value: cookie,
				},
			},
		},
	}); err != nil {
		return err
	}

	lg.Info(shared.EventOperatorLoggedIn)

	// создание нового сообщения
	if c, err := s.db.Chat.
		Create().
		SetMessage(username + " logged in").
		Save(ctx); err != nil {
		lg.Warn(shared.ErrorSaveChatMessage, zap.Error(err))
	} else {
		// нотификация всех подписчиков о создании нового сообщения
		go pools.Pool.Chat.Send(&operatorv1.ChatResponse{
			CreatedAt: timestamppb.New(c.CreatedAt),
			From:      defaults.ChatSrvFrom,
			Message:   c.Message,
		})
	}

	// тикер каждые N секунд, чтобы проверять работоспособность клиента оператора
	ticker := time.NewTicker(constants.GrpcOperatorHealthCheckTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := stream.Send(&operatorv1.HelloResponse{
				Response: &operatorv1.HelloResponse_Empty{},
			}); err != nil {
				lg.Warn(shared.EventOperatorLoggedOutWithError, zap.Error(err))
				return status.Error(codes.Canceled, shared.ErrorSend)
			}
		case <-ctx.Done():
			// клиент отключился
			lg.Info(shared.EventOperatorLoggedOut)
			return nil
		}
	}
}

// Подписка оператором на получение обновлений по зарегистрированным листенерам.
func (s *server) SubscribeListeners(req *operatorv1.SubscribeListenersRequest, stream operatorv1.OperatorService_SubscribeListenersServer) error {
	ctx := stream.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SubscribeListeners").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// проверяем, что оператора нет в листенеровой мапе
	if pools.Pool.Listeners.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected, zap.String("username", username))
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	// сохраняем оператора в мапу для отдачи списка листенеров
	pools.Pool.Listeners.Add(cookie, username, stream)

	defer func() {
		// удаление из листенеровой мапы
		pools.Pool.Listeners.Remove(cookie)
	}()

	lg.Info(shared.EventOperatorSubscribed)

	// получаем все листнеры из БД
	ls, err := s.db.Listener.
		Query().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryListenersFromDB, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// разбивем массив объектов на чанки и отправляем их
	for _, chunk := range utils.ChunkBy(ls, constants.MaxObjectChunks) {
		if err = stream.Send(&operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Listeners{
				Listeners: pools.ToListenersResponse(chunk),
			},
		}); err != nil {
			lg.Error(shared.ErrorSendListener, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorSend)
		}
	}

	// получаем данные подписки
	val := pools.Pool.Listeners.Get(cookie)
	if val == nil {
		lg.Error(shared.ErrorGetSubscriptionData)
		return status.Error(codes.Internal, shared.ErrorGetSubscriptionData)
	}

	for {
		select {
		case <-val.IsDisconnect():
			lg.Info(shared.EventOperatorUnsubscribedLoggedOut)
			return nil
		case err = <-val.Error():
			lg.Error(shared.ErrorDuringSubscription, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorDuringSubscription)
		case <-ctx.Done():
			lg.Info(shared.EventOperatorUnsubscribed)
			return nil
		}
	}
}

// Подписка оператором на получение обновлений по зарегистрированным биконам.
func (s *server) SubscribeBeacons(req *operatorv1.SubscribeBeaconsRequest, stream operatorv1.OperatorService_SubscribeBeaconsServer) error {
	ctx := stream.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SubscribeBeacons").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// проверяем, что оператора нет в биконовой мапе
	if pools.Pool.Beacons.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected)
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	// сохраняем оператора в мапу для отдачи списка биконов
	pools.Pool.Beacons.Add(cookie, username, stream)

	defer func() {
		// удаление из биконовой мапы
		pools.Pool.Beacons.Remove(cookie)
	}()

	lg.Info(shared.EventOperatorSubscribed)

	// получаем все биконы из БД
	bs, err := s.db.Beacon.
		Query().
		WithListener().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryBeaconsFromDB, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// разбиваем массив объектов на чанки и отправляем
	for _, chunk := range utils.ChunkBy(bs, constants.MaxObjectChunks) {
		if err = stream.Send(&operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Beacons{
				Beacons: pools.ToBeaconsResponse(chunk),
			},
		}); err != nil {
			lg.Error(shared.ErrorSendListener, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorSend)
		}
	}

	// получаем данные подписки
	val := pools.Pool.Beacons.Get(cookie)
	if val == nil {
		lg.Error(shared.ErrorGetSubscriptionData)
		return status.Error(codes.Internal, shared.ErrorGetSubscriptionData)
	}

	for {
		select {
		case <-val.IsDisconnect():
			lg.Info(shared.EventOperatorUnsubscribedLoggedOut)
			return nil
		case err = <-val.Error():
			lg.Error(shared.ErrorDuringSubscription, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorDuringSubscription)
		case <-ctx.Done():
			lg.Info(shared.EventOperatorUnsubscribed)
			return nil
		}
	}
}

// Подписка оператором на получение обновлений по зарегистрированным биконам.
func (s *server) SubscribeOperators(req *operatorv1.SubscribeOperatorsRequest, stream operatorv1.OperatorService_SubscribeOperatorsServer) error {
	ctx := stream.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SubscribeOperators").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// проверяем, что оператора нет в операторной мапе
	if pools.Pool.Operators.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected)
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	// сохраняем оператора в мапу для отдачи списка операторов
	pools.Pool.Operators.Add(cookie, username, stream)

	defer func() {
		// удаление из операторной мапы
		pools.Pool.Operators.Remove(cookie)
	}()

	lg.Info(shared.EventOperatorSubscribed)

	// получаем всех операторов из БД
	os, err := s.db.Operator.
		Query().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryOperatorsFromDB, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// разбиваем массив объектов на чанки и отправляем
	for _, chunk := range utils.ChunkBy(os, constants.MaxObjectChunks) {
		if err = stream.Send(&operatorv1.SubscribeOperatorsResponse{
			Response: &operatorv1.SubscribeOperatorsResponse_Operators{
				Operators: pools.ToOperatorsResponse(chunk),
			},
		}); err != nil {
			lg.Error(shared.ErrorSendListener, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorSend)
		}
	}

	// получаем данные подписки
	val := pools.Pool.Operators.Get(cookie)
	if val == nil {
		lg.Error(shared.ErrorGetSubscriptionData)
		return status.Error(codes.Internal, shared.ErrorGetSubscriptionData)
	}

	for {
		select {
		case <-val.IsDisconnect():
			lg.Info(shared.EventOperatorUnsubscribedLoggedOut)
			return nil
		case err = <-val.Error():
			lg.Error(shared.ErrorDuringSubscription, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorDuringSubscription)
		case <-ctx.Done():
			lg.Info(shared.EventOperatorUnsubscribed)
			return nil
		}
	}
}

// Подписка оператором на получение обновлений по сообщениям в чате.
func (s *server) SubscribeChat(req *operatorv1.SubscribeChatRequest, stream operatorv1.OperatorService_SubscribeChatServer) error {
	ctx := stream.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SubscribeChat").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// проверяем, что оператора нет в операторной мапе
	if pools.Pool.Chat.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected)
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	// сохраняем оператора в мапу для отдачи сообщений в чате
	pools.Pool.Chat.Add(cookie, username, stream)

	defer func() {
		// удаление из операторной мапы
		pools.Pool.Chat.Remove(cookie)
	}()

	lg.Info(shared.EventOperatorSubscribed)

	// получаем все сообщения чата из БД
	ms, err := s.db.Chat.
		Query().
		WithOperator().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryChatMessagesFromDB, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// разбиваем массив объектов на чанки и отправляем
	for _, chunk := range utils.ChunkBy(ms, constants.MaxObjectChunks) {
		if err = stream.Send(&operatorv1.SubscribeChatResponse{
			Response: &operatorv1.SubscribeChatResponse_Messages{
				Messages: pools.ToChatMessagesResponse(chunk),
			},
		}); err != nil {
			lg.Error(shared.ErrorSendListener, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorSend)
		}
	}

	// получаем данные подписки
	val := pools.Pool.Chat.Get(cookie)
	if val == nil {
		lg.Error(shared.ErrorGetSubscriptionData)
		return status.Error(codes.Internal, shared.ErrorGetSubscriptionData)
	}

	for {
		select {
		case <-val.IsDisconnect():
			lg.Info(shared.EventOperatorUnsubscribedLoggedOut)
			return nil
		case err = <-val.Error():
			lg.Error(shared.ErrorDuringSubscription, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorDuringSubscription)
		case <-ctx.Done():
			lg.Info(shared.EventOperatorUnsubscribed)
			return nil
		}
	}
}

// Подписка оператором на получение обновлений по сообщениям в чате.
func (s *server) SubscribeCredentials(req *operatorv1.SubscribeCredentialsRequest, stream operatorv1.OperatorService_SubscribeCredentialsServer) error {
	ctx := stream.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SubscribeCredentials").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// проверяем, что оператора нет в операторной мапе
	if pools.Pool.Credentials.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected)
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	// сохраняем оператора в мапу для отдачи кределов в чате
	pools.Pool.Credentials.Add(cookie, username, stream)

	defer func() {
		// удаление из операторной мапы
		pools.Pool.Credentials.Remove(cookie)
	}()

	lg.Info(shared.EventOperatorSubscribed)

	// получаем все кредлы из БД
	cs, err := s.db.Credential.
		Query().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryCredentialsFromDB, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// разбиваем массив объектов на чанки и отправляем
	for _, chunk := range utils.ChunkBy(cs, constants.MaxObjectChunks) {
		if err = stream.Send(&operatorv1.SubscribeCredentialsResponse{
			Response: &operatorv1.SubscribeCredentialsResponse_Credentials{
				Credentials: pools.ToCredentialsResponse(chunk),
			},
		}); err != nil {
			lg.Error(shared.ErrorSendListener, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorSend)
		}
	}

	// получаем данные подписки
	val := pools.Pool.Credentials.Get(cookie)
	if val == nil {
		lg.Error(shared.ErrorGetSubscriptionData)
		return status.Error(codes.Internal, shared.ErrorGetSubscriptionData)
	}

	for {
		select {
		case <-val.IsDisconnect():
			lg.Info(shared.EventOperatorUnsubscribedLoggedOut)
			return nil
		case err = <-val.Error():
			lg.Error(shared.ErrorDuringSubscription, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorDuringSubscription)
		case <-stream.Context().Done():
			lg.Info(shared.EventOperatorUnsubscribed)
			return nil
		}
	}
}

// Установка цвета для одного набора кредов.
func (s *server) SetCredentialColor(ctx context.Context, req *operatorv1.SetCredentialColorRequest) (*operatorv1.SetCredentialColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetCredentialColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем креды
	v, err := s.db.Credential.
		Query().
		Where(credential.ID(int(req.GetCid()))).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryCredentialFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// обновляем цвет
	if v, err = v.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetColorForCredential, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене цвета для кредов
	go pools.Pool.Credentials.Send(&operatorv1.CredentialColorResponse{
		Cid:   int64(v.ID),
		Color: wrapperspb.UInt32(v.Color),
	})

	return &operatorv1.SetCredentialColorResponse{}, nil
}

// Установка цвета для массива из наборов с кредами.
func (s *server) SetCredentialsColor(ctx context.Context, req *operatorv1.SetCredentialsColorRequest) (*operatorv1.SetCredentialsColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetCredentialsColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, id := range req.GetCids() {
		// получаем креды
		v, err := s.db.Credential.
			Query().
			Where(credential.ID(int(id))).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryCredentialFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем цвет
		if v, err = v.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetColorForCredential, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене цвета для кредов
		go pools.Pool.Credentials.Send(&operatorv1.CredentialColorResponse{
			Cid:   int64(v.ID),
			Color: wrapperspb.UInt32(v.Color),
		})
	}

	return &operatorv1.SetCredentialsColorResponse{}, nil
}

// Установка заметки для одного набора кредов.
func (s *server) SetCredentialNote(ctx context.Context, req *operatorv1.SetCredentialNoteRequest) (*operatorv1.SetCredentialNoteResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetCredentialNote").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем креды
	v, err := s.db.Credential.
		Query().
		Where(credential.ID(int(req.GetCid()))).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryCredentialFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// обновляем заметку
	if v, err = v.Update().SetNote(req.GetNote()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetNoteForCredential, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене заметки для кредов
	go pools.Pool.Credentials.Send(pools.ToCredentialNoteResponse(v))

	return &operatorv1.SetCredentialNoteResponse{}, nil
}

// Установка заметки для массива из набора кредов.
func (s *server) SetCredentialsNote(ctx context.Context, req *operatorv1.SetCredentialsNoteRequest) (*operatorv1.SetCredentialsNoteResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetCredentialsNote").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, id := range req.GetCids() {
		// получаем креды
		v, err := s.db.Credential.
			Query().
			Where(credential.ID(int(id))).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryCredentialFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем заметку
		if v, err = v.Update().SetNote(req.GetNote()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetNoteForCredential, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене заметки для кредов
		go pools.Pool.Credentials.Send(pools.ToCredentialNoteResponse(v))
	}

	return &operatorv1.SetCredentialsNoteResponse{}, nil
}

// Изменение цвета для листенера.
func (s *server) SetListenerColor(ctx context.Context, req *operatorv1.SetListenerColorRequest) (*operatorv1.SetListenerColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetListenerColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем листенер
	l, err := s.db.Listener.
		Query().
		Where(listener.ID(int(req.GetLid()))).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// обновляем цвет
	if l, err = l.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetColorForListener, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене цвета для листенера
	go pools.Pool.Listeners.Send(&operatorv1.ListenerColorResponse{
		Lid:   int64(l.ID),
		Color: wrapperspb.UInt32(l.Color),
	})

	return &operatorv1.SetListenerColorResponse{}, nil
}

// Изменение цвета массива из листенеров.
func (s *server) SetListenersColor(ctx context.Context, req *operatorv1.SetListenersColorRequest) (*operatorv1.SetListenersColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetListenersColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, id := range req.GetLids() {
		// получаем листенер
		l, err := s.db.Listener.
			Query().
			Where(listener.ID(int(id))).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем цвет
		if l, err = l.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetColorForListener, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене цвета для листенера
		go pools.Pool.Listeners.Send(&operatorv1.ListenerColorResponse{
			Lid:   int64(l.ID),
			Color: wrapperspb.UInt32(l.Color),
		})
	}

	return &operatorv1.SetListenersColorResponse{}, nil
}

// Установка заметки на листенер.
func (s *server) SetListenerNote(ctx context.Context, req *operatorv1.SetListenerNoteRequest) (*operatorv1.SetListenerNoteResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetListenerNote").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем листенер
	l, err := s.db.Listener.
		Query().
		Where(listener.ID(int(req.GetLid()))).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// обновляем заметку
	if l, err = l.Update().SetNote(req.GetNote()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetNoteForListener, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене заметки для листенера
	go pools.Pool.Listeners.Send(pools.ToListenerNoteResponse(l))

	return &operatorv1.SetListenerNoteResponse{}, nil
}

// Установка заметки на массив из листенеров.
func (s *server) SetListenersNote(ctx context.Context, req *operatorv1.SetListenersNoteRequest) (*operatorv1.SetListenersNoteResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetListenersNote").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, id := range req.GetLids() {
		// получаем листенер
		l, err := s.db.Listener.
			Query().
			Where(listener.ID(int(id))).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем заметку
		if l, err = l.Update().SetNote(req.GetNote()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetNoteForListener, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене заметки для листенера
		go pools.Pool.Listeners.Send(pools.ToListenerNoteResponse(l))
	}

	return &operatorv1.SetListenersNoteResponse{}, nil
}

// Изменение цвета для бикона.
func (s *server) SetBeaconColor(ctx context.Context, req *operatorv1.SetBeaconColorRequest) (*operatorv1.SetBeaconColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetBeaconColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем бикон
	b, err := s.db.Beacon.
		Query().
		Where(beacon.Bid(req.GetBid())).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryBeaconFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// обновляем цвет
	if b, err = b.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetColorForBeacon, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене цвета для бикона
	go pools.Pool.Beacons.Send(&operatorv1.BeaconColorResponse{
		Bid:   b.Bid,
		Color: wrapperspb.UInt32(b.Color),
	})

	return &operatorv1.SetBeaconColorResponse{}, nil
}

// Изменение цвета для массива из биконов.
func (s *server) SetBeaconsColor(ctx context.Context, req *operatorv1.SetBeaconsColorRequest) (*operatorv1.SetBeaconsColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetBeaconsColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, bid := range req.GetBids() {
		// получаем бикон
		b, err := s.db.Beacon.
			Query().
			Where(beacon.Bid(bid)).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryBeaconFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем цвет
		if b, err = b.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetColorForBeacon, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене цвета для бикона
		go pools.Pool.Beacons.Send(&operatorv1.BeaconColorResponse{
			Bid:   b.Bid,
			Color: wrapperspb.UInt32(b.Color),
		})
	}

	return &operatorv1.SetBeaconsColorResponse{}, nil
}

// Установка заметки для бикона.
func (s *server) SetBeaconNote(ctx context.Context, req *operatorv1.SetBeaconNoteRequest) (*operatorv1.SetBeaconNoteResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetBeaconNote").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем бикон
	b, err := s.db.Beacon.
		Query().
		Where(beacon.Bid(req.GetBid())).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryBeaconFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorQueryBeaconFromDB)
	}
	// обновляем заметку
	if b, err = b.Update().SetNote(req.GetNote()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetNoteForBeacon, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене заметки для бикона
	go pools.Pool.Beacons.Send(&operatorv1.BeaconNoteResponse{
		Bid:  b.Bid,
		Note: wrapperspb.String(b.Note),
	})

	return &operatorv1.SetBeaconNoteResponse{}, nil
}

// Установка заметки для массива из биконов.
func (s *server) SetBeaconsNote(ctx context.Context, req *operatorv1.SetBeaconsNoteRequest) (*operatorv1.SetBeaconsNoteResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetBeaconsNote").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, bid := range req.GetBids() {
		// получаем бикон
		b, err := s.db.Beacon.
			Query().
			Where(beacon.Bid(bid)).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryBeaconsFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем заметку
		if b, err = b.Update().SetNote(req.GetNote()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetNoteForBeacon, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене заметки для бикона
		go pools.Pool.Beacons.Send(&operatorv1.BeaconNoteResponse{
			Bid:  b.Bid,
			Note: wrapperspb.String(b.Note),
		})
	}

	return &operatorv1.SetBeaconsNoteResponse{}, nil
}

// Изменение цвета оператора.
func (s *server) SetOperatorColor(ctx context.Context, req *operatorv1.SetOperatorColorRequest) (*operatorv1.SetOperatorColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetOperatorColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем оператора
	o, err := s.db.Operator.
		Query().
		Where(operator.Username(req.GetUsername())).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// обновляем цвет
	if o, err = o.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
		lg.Error(shared.ErrorSetColorForOperator, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// нотификация всех подписчиков о смене цвета для оператора
	go pools.Pool.Operators.Send(&operatorv1.OperatorColorResponse{
		Username: o.Username,
		Color:    wrapperspb.UInt32(o.Color),
	})

	return &operatorv1.SetOperatorColorResponse{}, nil
}

// Изменение цвета массива из операторов.
func (s *server) SetOperatorsColor(ctx context.Context, req *operatorv1.SetOperatorsColorRequest) (*operatorv1.SetOperatorsColorResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SetOperatorColor").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	for _, u := range req.GetUsernames() {
		// получаем оператора
		o, err := s.db.Operator.
			Query().
			Where(operator.Username(u)).
			Only(ctx)
		if err != nil {
			lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// обновляем цвет
		if o, err = o.Update().SetColor(req.GetColor()).Save(ctx); err != nil {
			lg.Error(shared.ErrorSetColorForOperator, zap.Error(err))
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
		// нотификация всех подписчиков о смене цвета для оператора
		go pools.Pool.Operators.Send(&operatorv1.OperatorColorResponse{
			Username: o.Username,
			Color:    wrapperspb.UInt32(o.Color),
		})
	}

	return &operatorv1.SetOperatorsColorResponse{}, nil
}

// Создание нового сообщение от оператора в чате.
func (s *server) NewChatMessage(ctx context.Context, req *operatorv1.NewChatMessageRequest) (*operatorv1.NewChatMessageResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("NewChatMessage").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// получаем оператора
	o, err := s.db.Operator.
		Query().
		Where(operator.Username(username)).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	// сохраняем новое сообщение
	c, err := s.db.Chat.
		Create().
		SetOperator(o).
		SetMessage(req.GetMessage()).
		Save(ctx)
	if err != nil {
		lg.Error(shared.ErrorSaveChatMessage, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// нотификация всех подписчиков о новом сообщении
	go pools.Pool.Chat.Send(&operatorv1.ChatResponse{
		CreatedAt: timestamppb.New(c.CreatedAt),
		From:      o.Username,
		Message:   c.Message,
	})

	return &operatorv1.NewChatMessageResponse{}, nil
}

// Создание новой связки кредов.
func (s *server) NewCredential(ctx context.Context, req *operatorv1.NewCredentialRequest) (*operatorv1.NewCredentialResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("NewCredential").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return nil, status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// начало заполнения структуры кредов
	c := s.db.Credential.Create()

	// username
	if req.GetUsername() != nil {
		if len(req.GetUsername().GetValue()) > defaults.CredentialUsernameMaxLength {
			c.SetUsername(req.GetUsername().GetValue()[:defaults.CredentialUsernameMaxLength])
		} else {
			c.SetUsername(req.GetUsername().GetValue())
		}
	}
	// password
	if req.GetPassword() != nil {
		if len(req.GetPassword().GetValue()) > defaults.CredentialSecretMaxLength {
			c.SetPassword(req.GetPassword().GetValue()[:defaults.CredentialSecretMaxLength])
		} else {
			c.SetPassword(req.GetPassword().GetValue())
		}
	}
	// realm
	if req.GetRealm() != nil {
		if len(req.GetRealm().GetValue()) > defaults.CredentialRealmMaxLength {
			c.SetRealm(req.GetRealm().GetValue()[:defaults.CredentialRealmMaxLength])
		} else {
			c.SetRealm(req.GetRealm().GetValue())
		}
	}
	// host
	if req.GetHost() != nil {
		if len(req.GetHost().GetValue()) > defaults.CredentialHostMaxLength {
			c.SetHost(req.GetHost().GetValue()[:defaults.CredentialHostMaxLength])
		} else {
			c.SetHost(req.GetHost().GetValue())
		}
	}

	c.SetColor(constants.DefaultColor)

	// сохраняем
	var cs *ent.Credential
	var err error
	if cs, err = c.Save(ctx); err != nil {
		lg.Error(shared.ErrorSaveCredential, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// рассылаем подписчикам новые креды
	go pools.Pool.Credentials.Send(pools.ToCredentialResponse(cs))

	return &operatorv1.NewCredentialResponse{}, nil
}

// Создание новой группы/тасков/сообщений для бикона.
func (s *server) NewGroup(ss operatorv1.OperatorService_NewGroupServer) error {
	ctx := ss.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("NewGroup").With(zap.String("username", username))

	// первый запрос - создание самой таск группы
	val, err := ss.Recv()
	if err != nil {
		lg.Error("receive group request", zap.Error(err))
		return status.Error(codes.Internal, "receive group request failed")
	}

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, val.GetCookie().GetValue()) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	if val.GetGroup() == nil {
		// если первый запрос не на создание таск группы -> дропаем
		lg.Warn("first message ss not request for task group creation")
		return status.Error(codes.InvalidArgument, "invalid initial message")
	}

	// получаем объект автора
	o, err := s.db.Operator.
		Query().
		Where(operator.Username(username)).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// получаем объект бикона
	b, err := s.db.Beacon.
		Query().
		Where(beacon.Bid(val.GetGroup().GetBid())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			lg.Warn(shared.ErrorUnknownBeacon, zap.Error(err))
		} else {
			lg.Error(shared.ErrorQueryBeaconFromDB, zap.Error(err))
		}
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// создаем новую таск группу
	g, err := s.db.Group.
		Create().
		SetBeacon(b).
		SetOperator(o).
		SetCmd(val.GetGroup().GetCmd()).
		SetVisible(val.GetGroup().GetVisible()).
		Save(ctx)
	if err != nil {
		lg.Error(shared.ErrorSaveGroup, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorDB)
	}
	// уведомление о новой группе
	go pools.Pool.Tasks.Send(&operatorv1.TasksGroupResponse{
		Gid:     int64(g.ID),
		Bid:     b.Bid,
		Cmd:     g.Cmd,
		Author:  o.Username,
		Created: timestamppb.New(g.CreatedAt),
		Visible: g.Visible,
	})

	for {
		msg, err := ss.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			lg.Error(shared.ErrorReceive, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorReceive)
		}

		cookie := msg.GetCookie().GetValue()

		// проверяем, что кука и username совпадают
		if !pools.Pool.Hello.Validate(username, cookie) {
			lg.Warn(shared.ErrorInvalidSessionCookie)
			return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
		}

		// сообщение в группу
		if msg.GetMessage() != nil {
			// сохраняем новое сообщение
			m, err := s.db.Message.
				Create().
				SetGroup(g).
				SetMessage(msg.GetMessage().GetMsg()).
				SetType(def.TaskMessage(msg.GetMessage().GetType())).
				Save(ctx)
			if err != nil {
				lg.Error(shared.ErrorSaveGroupMessage, zap.Error(err))
				return status.Error(codes.Internal, shared.ErrorDB)
			}
			// рассылаем всем подписчикам
			go pools.Pool.Tasks.Send(&operatorv1.TasksMessageResponse{
				Gid:     int64(g.ID),
				Bid:     b.Bid,
				Mid:     int64(m.ID),
				Type:    uint32(m.Type),
				Message: m.Message,
				Created: timestamppb.New(m.CreatedAt),
			})
		}

		// создание нового таска
		if msg.GetTask() != nil {
			var raw []byte

			c := def.Capability(msg.GetTask().GetCap())
			switch c {
			case def.CAP_SLEEP:
				raw, err = c.Marshal(msg.GetTask().GetSleep())
			case def.CAP_LS:
				raw, err = c.Marshal(msg.GetTask().GetLs())
			case def.CAP_PWD:
				raw, err = c.Marshal(msg.GetTask().GetPwd())
			case def.CAP_CD:
				raw, err = c.Marshal(msg.GetTask().GetCd())
			case def.CAP_WHOAMI:
				raw, err = c.Marshal(msg.GetTask().GetWhoami())
			case def.CAP_PS:
				raw, err = c.Marshal(msg.GetTask().GetPs())
			case def.CAP_CAT:
				raw, err = c.Marshal(msg.GetTask().GetCat())
			case def.CAP_EXEC:
				raw, err = c.Marshal(msg.GetTask().GetExec())
			case def.CAP_CP:
				raw, err = c.Marshal(msg.GetTask().GetCp())
			case def.CAP_JOBS:
				raw, err = c.Marshal(msg.GetTask().GetJobs())
			case def.CAP_JOBKILL:
				raw, err = c.Marshal(msg.GetTask().GetJobkill())
			case def.CAP_KILL:
				raw, err = c.Marshal(msg.GetTask().GetKill())
			case def.CAP_MV:
				raw, err = c.Marshal(msg.GetTask().GetMv())
			case def.CAP_MKDIR:
				raw, err = c.Marshal(msg.GetTask().GetMkdir())
			case def.CAP_RM:
				raw, err = c.Marshal(msg.GetTask().GetRm())
			case def.CAP_EXEC_ASSEMBLY:
				raw, err = c.Marshal(msg.GetTask().GetExecAssembly())
			case def.CAP_SHELLCODE_INJECTION:
				raw, err = c.Marshal(msg.GetTask().GetShellcodeInjection())
			case def.CAP_DOWNLOAD:
				raw, err = c.Marshal(msg.GetTask().GetDownload())
			case def.CAP_UPLOAD:
				raw, err = c.Marshal(msg.GetTask().GetUpload())
			case def.CAP_PAUSE:
				raw, err = c.Marshal(msg.GetTask().GetPause())
			case def.CAP_DESTRUCT:
				raw, err = c.Marshal(msg.GetTask().GetDestruct())
			case def.CAP_EXEC_DETACH:
				raw, err = c.Marshal(msg.GetTask().GetExecDetach())
			case def.CAP_SHELL:
				raw, err = c.Marshal(msg.GetTask().GetShell())
			case def.CAP_PPID:
				raw, err = c.Marshal(msg.GetTask().GetPpid())
			case def.CAP_EXIT:
				raw, err = c.Marshal(msg.GetTask().GetExit())
			default:
				err = fmt.Errorf("unknown capability %d", c)
			}

			if err != nil {
				// если не смогли отмаршалить данные -> дропаем
				lg.Warn(shared.ErrorMarshalCapability, zap.Error(err))
				return status.Error(codes.InvalidArgument, shared.ErrorMarshalCapability)
			}

			// кладем аргументы в блоббер
			var blob *ent.Blobber
			h := utils.CalcHash(raw)
			if blob, err = s.db.Blobber.Query().Where(blobber.Hash(h)).Only(ctx); err != nil {
				if ent.IsNotFound(err) {
					// создание блоба
					if blob, err = s.db.Blobber.
						Create().
						SetBlob(raw).
						SetHash(h).
						SetSize(len(raw)).
						Save(ctx); err != nil {
						lg.Error(shared.ErrorSaveBlob, zap.Error(err))
						return status.Error(codes.Internal, shared.ErrorDB)
					}
				} else {
					lg.Error(shared.ErrorQueryBlobFromDB, zap.Error(err))
					return status.Error(codes.Internal, shared.ErrorDB)
				}
			}
			// создание таска
			t, err := s.db.Task.
				Create().
				SetBeacon(b).
				SetGroup(g).
				SetBlobberArgs(blob).
				SetCap(c).
				SetStatus(def.StatusNew).
				Save(ctx)
			if err != nil {
				lg.Error(shared.ErrorSaveGroupTask, zap.Error(err))
				return status.Error(codes.Internal, shared.ErrorDB)
			}
			// нотификация подписчиков о создании нового таска
			go pools.Pool.Tasks.Send(&operatorv1.TasksResponse{
				Gid:       int64(g.ID),
				Tid:       int64(t.ID),
				Bid:       b.Bid,
				Status:    uint32(t.Status),
				Created:   timestamppb.New(t.CreatedAt),
				OutputBig: false,
			})
		}
	}

	return nil
}

// Подписка оператора на получение обновлений по таскам для биконов.
func (s *server) SubscribeTasks(ss operatorv1.OperatorService_SubscribeTasksServer) error {
	ctx := ss.Context()

	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("SubscribeTasks").With(zap.String("username", username))

	// первый запрос должен быть hello
	h, err := ss.Recv()
	if err != nil {
		lg.Error(shared.ErrorGetFirstHelloMsg, zap.Error(err))
		return status.Error(codes.Internal, shared.ErrorGetFirstHelloMsg)
	}
	if h.GetHello() == nil {
		lg.Error(shared.ErrorFirstHelloMsg)
		return status.Error(codes.InvalidArgument, shared.ErrorFirstHelloMsg)
	}

	cookie := h.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn(shared.ErrorInvalidSessionCookie)
		return status.Error(codes.PermissionDenied, shared.ErrorInvalidSessionCookie)
	}

	// проверяем, что оператора нет в тасковой мапе
	if pools.Pool.Tasks.Exists(username) {
		lg.Warn(shared.ErrorOperatorAlreadyConnected)
		return status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyConnected)
	}

	// сохраняем оператора в мапу для отдачи обновлений по таскам
	pools.Pool.Tasks.Add(cookie, username, ss)

	defer func() {
		// удаление из тасковой мапы
		pools.Pool.Tasks.Remove(cookie)
	}()

	lg.Info(shared.EventOperatorSubscribed)

	val := pools.Pool.Tasks.Get(cookie)
	if val == nil {
		lg.Error(shared.ErrorGetSubscriptionData)
		return status.Error(codes.Internal, shared.ErrorGetSubscriptionData)
	}

	// получаем оператора
	o, err := s.db.Operator.
		Query().
		Where(operator.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryOperatorFromDB)
		return status.Error(codes.Internal, shared.ErrorQueryOperatorFromDB)
	}

	gw, subCtx := errgroup.WithContext(ctx)

	gw.Go(func() error {
		for {
			msg, err := ss.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}

			cookie = msg.GetCookie().GetValue()

			// валидируем, что кука валидная
			if !pools.Pool.Hello.Validate(username, cookie) {
				return errors.New(shared.ErrorInvalidSessionCookie)
			}

			// добавялем бикона в поллинг
			if msg.GetStart() != nil {
				// отдаем все группы для бикона
				b, err := s.db.Beacon.
					Query().
					Where(beacon.BidEQ(msg.GetStart().GetBid())).
					Only(ctx)
				if err != nil {
					return errs.Wrap(err, shared.ErrorQueryBeaconFromDB)
				}
				gs, err := s.db.Group.
					Query().
					WithOperator().
					Where(group.BidEQ(b.ID)).
					// эта логика нужна, чтобы не отдавать invisible группы другим операторам
					Where(group.Not(group.And(
						group.AuthorNEQ(o.ID),
						group.VisibleEQ(false),
					))).
					// чтобы делать выборки даже по удаленным операторам
					All(ctx)
				if err != nil {
					return errs.Wrap(err, shared.ErrorQueryTaskGroupsFromDB)
				}
				for _, g := range gs {
					gOperator, err := g.Edges.OperatorOrErr()
					if err != nil {
						lg.Warn("unable load operator for group", zap.Int("gid", g.ID))
						continue
					}
					if err = ss.Send(&operatorv1.SubscribeTasksResponse{
						Type: &operatorv1.SubscribeTasksResponse_Group{
							Group: &operatorv1.TasksGroupResponse{
								Gid:     int64(g.ID),
								Bid:     b.Bid,
								Cmd:     g.Cmd,
								Author:  gOperator.Username,
								Created: timestamppb.New(g.CreatedAt),
								Visible: g.Visible,
							},
						},
					}); err != nil {
						return errs.Wrap(err, shared.ErrorSendGroup)
					}
					// получаем и отправляем связанные с группой сообщения
					msgs, err := s.db.Message.
						Query().
						Where(message.GidEQ(g.ID)).
						All(ctx)
					if err != nil {
						return errs.Wrap(err, shared.ErrorQueryTaskGroupMessagesFromDB)
					}
					for _, m := range msgs {
						if err = ss.Send(&operatorv1.SubscribeTasksResponse{
							Type: &operatorv1.SubscribeTasksResponse_Message{
								Message: &operatorv1.TasksMessageResponse{
									Gid:     int64(g.ID),
									Mid:     int64(m.ID),
									Bid:     b.Bid,
									Type:    uint32(m.Type),
									Message: m.Message,
									Created: timestamppb.New(m.CreatedAt),
								},
							},
						}); err != nil {
							return errs.Wrap(err, shared.ErrorSendTaskMessage)
						}
					}
					// получаем и отправляем связанные с группой таски
					ts, err := s.db.Task.
						Query().
						WithBlobberOutput().
						Where(task.GidEQ(g.ID)).
						All(ctx)
					if err != nil {
						return errs.Wrap(err, shared.ErrorQueryTasksFromDB)
					}
					for _, t := range ts {
						nt := &operatorv1.TasksResponse{
							Gid:       int64(g.ID),
							Tid:       int64(t.ID),
							Bid:       b.Bid,
							Status:    uint32(t.Status),
							OutputBig: t.OutputBig,
							Created:   timestamppb.New(t.CreatedAt),
						}
						// если размер output таска меньше N -> подкладываем его
						blob, err := t.Edges.BlobberOutputOrErr()
						if err != nil {
							if !ent.IsNotFound(err) {
								// если блоб не найден - значит он не существует еще для таска
								return errs.Wrap(err, shared.ErrorQueryBlobFromDB)
							}
							nt.OutputLen = 0
						} else {
							nt.OutputLen = int64(blob.Size)
							if !t.OutputBig {
								nt.Output = wrapperspb.Bytes(blob.Blob)
							}
						}
						if err = ss.Send(&operatorv1.SubscribeTasksResponse{
							Type: &operatorv1.SubscribeTasksResponse_Task{
								Task: nt,
							},
						}); err != nil {
							return errs.Wrap(err, "send task from group to operator")
						}
					}
				}
				// добавляем бикон в поллинг
				pools.Pool.Tasks.AddBeacon(cookie, msg.GetStart().GetBid())
			}

			// убираем бикона из поллинга
			if msg.GetStop() != nil {
				pools.Pool.Tasks.DeleteBeacon(cookie, msg.GetStop().GetBid())
			}
		}
	})

	for {
		select {
		case <-val.IsDisconnect():
			lg.Info(shared.EventOperatorUnsubscribedLoggedOut)
			return nil
		case err = <-val.Error():
			lg.Error(shared.ErrorDuringSubscription, zap.Error(err))
			return status.Error(codes.Internal, shared.ErrorDuringSubscription)
		case <-subCtx.Done():
			if err = gw.Wait(); err != nil {
				lg.Error("error during receiving", zap.Error(err))
				return status.Error(codes.Internal, "something went wrong")
			}
			return nil
		case <-ctx.Done():
			lg.Info(shared.EventOperatorUnsubscribed)
			return nil
		}
	}
}

// Перевод статуса тасок для бикона из NEW в CANCELLED, которые созданы оператором ранее.
func (s *server) CancelTasks(ctx context.Context, req *operatorv1.CancelTasksRequest) (*operatorv1.CancelTasksResponse, error) {
	username := grpcauth.OperatorFromCtx(ctx)
	lg := s.lg.Named("CancelTasks").With(zap.String("username", username))
	cookie := req.GetCookie().GetValue()

	// проверяем, что кука и username совпадают
	if !pools.Pool.Hello.Validate(username, cookie) {
		lg.Warn("specified invalid session cookie")
		return nil, status.Error(codes.PermissionDenied, "invalid session")
	}

	// получаем бикон
	b, err := s.db.Beacon.
		Query().
		Where(beacon.BidEQ(req.GetBid())).
		Only(ctx)
	if err != nil {
		lg.Error("query beacon from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable query beacon")
	}

	// получаем опертора
	o, err := s.db.Operator.
		Query().
		Where(operator.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		lg.Error("query operator from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable query operator")
	}

	// получаем таски
	ts, err := s.db.Task.
		Query().
		WithGroup(func(q *ent.GroupQuery) {
			q.Where(group.AuthorEQ(o.ID))
		}).
		Order(task.ByCreatedAt()).
		Where(task.StatusEQ(def.StatusNew)).
		Where(task.BidEQ(b.ID)).
		All(ctx)
	if err != nil {
		lg.Error("query tasks from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable query tasks")
	}
	for _, t := range ts {
		t, err = t.Update().
			SetStatus(def.StatusCancelled).
			SetDoneAt(time.Now()).
			Save(ctx)
		if err != nil {
			lg.Warn("update task status", zap.Error(err))
			continue
		}
		// оповещение подписчиков об отмене таска
		go pools.Pool.Tasks.Send(&operatorv1.TasksStatusResponse{
			Bid:    b.Bid,
			Tid:    int64(t.ID),
			Gid:    int64(t.Gid),
			Status: uint32(t.Status),
		})
	}
	return &operatorv1.CancelTasksResponse{}, nil
}
