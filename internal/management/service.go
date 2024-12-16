package management

import (
	"context"

	managementv1 "github.com/c2micro/c2mshr/proto/gen/management/v1"
	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/c2micro/c2msrv/internal/defaults"
	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/ent/listener"
	"github.com/c2micro/c2msrv/internal/ent/operator"
	"github.com/c2micro/c2msrv/internal/pools"
	"github.com/c2micro/c2msrv/internal/shared"
	"github.com/c2micro/c2msrv/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type server struct {
	managementv1.UnimplementedManagementServiceServer
	db *ent.Client
	lg *zap.Logger
}

// Получение списка операторов
func (s *server) GetOperators(ctx context.Context, req *managementv1.GetOperatorsRequest) (*managementv1.GetOperatorsResponse, error) {
	lg := s.lg.Named("GetOperators")

	// получение списка операторов
	operators, err := s.db.Operator.
		Query().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryOperatorsFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	var temp []*managementv1.Operator
	for _, operator := range operators {
		temp = append(temp, &managementv1.Operator{
			Username: operator.Username,
			Token:    wrapperspb.String(operator.Token),
			Last:     timestamppb.New(operator.Last),
		})
	}

	return &managementv1.GetOperatorsResponse{Operators: temp}, nil
}

// Создание нового оператора
func (s *server) NewOperator(ctx context.Context, req *managementv1.NewOperatorRequest) (*managementv1.NewOperatorResponse, error) {
	lg := s.lg.Named("NewOperator")

	// проверяем, что оператора нет в БД
	_, err := s.db.Operator.Query().Where(operator.Username(req.GetUsername())).Only(ctx)
	if err == nil {
		lg.Warn(shared.ErrorOperatorAlreadyExists, zap.String("username", req.GetUsername()))
		return nil, status.Error(codes.AlreadyExists, shared.ErrorOperatorAlreadyExists)
	} else if !ent.IsNotFound(err) {
		lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// создаем оператора
	operator, err := s.db.Operator.
		Create().
		SetUsername(req.GetUsername()).
		SetColor(defaults.DefaultColor).
		SetToken(utils.RandString(32)).
		Save(ctx)
	if err != nil {
		lg.Error(shared.ErrorSaveOperator, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// нотификация подписчиков о создании нового оператора
	go pools.Pool.Operators.Send(&operatorv1.OperatorResponse{
		Username: operator.Username,
		Last:     timestamppb.New(operator.Last),
		Color:    wrapperspb.UInt32(operator.Color),
	})

	return &managementv1.NewOperatorResponse{Operator: &managementv1.Operator{
		Username: operator.Username,
		Token:    wrapperspb.String(operator.Token),
		Last:     timestamppb.New(operator.Last),
	}}, nil
}

// Отзыв access токена оператора
func (s *server) RevokeOperator(ctx context.Context, req *managementv1.RevokeOperatorRequest) (*managementv1.RevokeOperatorResponse, error) {
	lg := s.lg.Named("RevokeOperator")

	// существует ли оператор с таким username
	if _, err := s.db.Operator.Query().Where(operator.Username(req.GetUsername())).Only(ctx); err != nil {
		if ent.IsNotFound(err) {
			lg.Warn(shared.ErrorUnknownOperator, zap.String("username", req.GetUsername()))
			return nil, status.Error(codes.InvalidArgument, shared.ErrorUnknownOperator)
		}
		lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	if _, err := s.db.Operator.
		Update().
		Where(operator.Username(req.GetUsername())).
		ClearToken().
		Save(ctx); err != nil {
		lg.Error(shared.ErrorUpdateOperator, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	return &managementv1.RevokeOperatorResponse{}, nil
}

// Регенерация access токена для оператора
func (s *server) RegenerateOperator(ctx context.Context, req *managementv1.RegenerateOperatorRequest) (*managementv1.RegenerateOperatorResponse, error) {
	lg := s.lg.Named("RegenerateOperator")

	// существует ли оператор с таким username
	o, err := s.db.Operator.Query().Where(operator.Username(req.GetUsername())).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			lg.Warn(shared.ErrorUnknownOperator, zap.String("username", req.GetUsername()))
			return nil, status.Error(codes.InvalidArgument, shared.ErrorUnknownOperator)
		}
		lg.Error(shared.ErrorQueryOperatorFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	token := utils.RandString(32)

	if _, err = s.db.Operator.
		Update().
		Where(operator.Username(req.GetUsername())).
		SetToken(token).
		Save(ctx); err != nil {
		lg.Error(shared.ErrorUpdateOperator, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	return &managementv1.RegenerateOperatorResponse{Operator: &managementv1.Operator{
		Username: o.Username,
		Token:    wrapperspb.String(token),
		Last:     timestamppb.New(o.Last),
	}}, nil
}

// Получение списка листенеров
func (s *server) GetListeners(ctx context.Context, req *managementv1.GetListenersRequest) (*managementv1.GetListenersResponse, error) {
	lg := s.lg.Named("GetListeners")

	// получение списка операторов
	listeners, err := s.db.Listener.
		Query().
		All(ctx)
	if err != nil {
		lg.Error(shared.ErrorQueryListenersFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	var temp []*managementv1.Listener
	for _, listener := range listeners {
		temp = append(temp, &managementv1.Listener{
			Lid:   int64(listener.ID),
			Name:  wrapperspb.String(listener.Name),
			Ip:    wrapperspb.String(listener.IP.String()),
			Port:  wrapperspb.UInt32(uint32(listener.Port)),
			Token: wrapperspb.String(listener.Token),
			Last:  timestamppb.New(listener.Last),
		})
	}

	return &managementv1.GetListenersResponse{Listeners: temp}, nil
}

// Создание нового листенера
func (s *server) NewListener(ctx context.Context, req *managementv1.NewListenerRequest) (*managementv1.NewListenerResponse, error) {
	lg := s.lg.Named("NewListener")

	// создаем листенер
	listener, err := s.db.Listener.
		Create().
		SetColor(defaults.DefaultColor).
		SetToken(utils.RandString(32)).
		Save(ctx)
	if err != nil {
		lg.Error(shared.ErrorSaveListener, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// нотификация подписчиков о создании нового листенера
	go pools.Pool.Listeners.Send(&operatorv1.ListenerResponse{
		Lid:   int64(listener.ID),
		Color: wrapperspb.UInt32(listener.Color),
	})

	return &managementv1.NewListenerResponse{Listener: &managementv1.Listener{
		Lid:   int64(listener.ID),
		Token: wrapperspb.String(listener.Token),
		Last:  timestamppb.New(listener.Last),
	}}, nil
}

// RevokeListener отзыв access токена листенера
func (s *server) RevokeListener(ctx context.Context, req *managementv1.RevokeListenerRequest) (*managementv1.RevokeListenerResponse, error) {
	lg := s.lg.Named("RevokeListener")

	// существует ли листенер с таким ID
	if _, err := s.db.Listener.Query().Where(listener.ID(int(req.GetLid()))).Only(ctx); err != nil {
		if ent.IsNotFound(err) {
			lg.Warn(shared.ErrorUnknownListener, zap.Int64("lid", req.GetLid()))
			return nil, status.Error(codes.InvalidArgument, shared.ErrorUnknownListener)
		}
		lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	if _, err := s.db.Listener.
		Update().
		Where(listener.ID(int(req.GetLid()))).
		ClearToken().
		Save(ctx); err != nil {
		lg.Error(shared.ErrorUpdateListener, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}
	return &managementv1.RevokeListenerResponse{}, nil
}

// Регенерация access токена для листенера
func (s *server) RegenerateListener(ctx context.Context, req *managementv1.RegenerateListenerRequest) (*managementv1.RegenerateListenerResponse, error) {
	lg := s.lg.Named("RegenerateListener")

	// существует ли листенер с таким ID
	l, err := s.db.Listener.Query().Where(listener.ID(int(req.GetLid()))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			lg.Warn(shared.ErrorUnknownListener, zap.Int64("lid", req.GetLid()))
			return nil, status.Error(codes.InvalidArgument, shared.ErrorUnknownListener)
		}
		lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	token := utils.RandString(32)

	if _, err := s.db.Listener.
		Update().
		Where(listener.ID(int(req.GetLid()))).
		SetToken(token).
		Save(ctx); err != nil {
		lg.Error(shared.ErrorUpdateListener, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	return &managementv1.RegenerateListenerResponse{Listener: &managementv1.Listener{
		Lid:   int64(l.ID),
		Name:  wrapperspb.String(l.Name),
		Ip:    wrapperspb.String(l.IP.String()),
		Port:  wrapperspb.UInt32(uint32(l.Port)),
		Token: wrapperspb.String(token),
		Last:  timestamppb.New(l.Last),
	}}, nil
}
