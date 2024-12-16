package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/lrita/cmap"
)

type SubscribeTasks struct {
	username   string
	disconnect chan struct{}
	bids       map[uint32]struct{}
	data       chan any
	err        chan error
	ss         operatorv1.OperatorService_SubscribeTasksServer
}

type PoolTasks struct {
	pool cmap.Map[string, *SubscribeTasks]
}

// Получение ошибки, возникшей при обработке сессии оператора.
func (s *SubscribeTasks) Error() chan error {
	return s.err
}

// Получение нотификации о небходимости отключения оператора от подписки.
func (s *SubscribeTasks) IsDisconnect() chan struct{} {
	return s.disconnect
}

// Добавление сессии оператора в мапу.
func (p *PoolTasks) Add(c, u string, s operatorv1.OperatorService_SubscribeTasksServer) {
	p.pool.Store(c, &SubscribeTasks{
		username:   u,
		disconnect: make(chan struct{}, 1),
		data:       make(chan any, 1),
		err:        make(chan error),
		bids:       make(map[uint32]struct{}),
		ss:         s,
	})
}

// Добавления бикона в поллинг для определенного оператора.
func (p *PoolTasks) AddBeacon(c string, bid uint32) {
	if s, ok := p.pool.Load(c); ok {
		s.bids[bid] = struct{}{}
	}
}

// Удаление бикона из поллинга для определенного оператора.
func (p *PoolTasks) DeleteBeacon(c string, bid uint32) {
	if s, ok := p.pool.Load(c); ok {
		delete(s.bids, bid)
	}
}

// Нотификация о необходимости удаления сессии оператора из мапы.
func (p *PoolTasks) Disconnect(c string) {
	v, ok := p.pool.Load(c)
	if ok {
		v.disconnect <- struct{}{}
	}
}

// Удаление сессии оператора из мапы.
func (p *PoolTasks) Remove(c string) {
	p.pool.Delete(c)
}

// Получение сессии оператора из мапы.
func (p *PoolTasks) Get(c string) *SubscribeTasks {
	t, _ := p.pool.Load(c)
	return t
}

// Отправка сообщения всем операторам, подписанным на топик с тасками.
// TODO - не отправлять результаты тасок и сообщения для невидимых таск-групп.
func (p *PoolTasks) Send(val any) {
	var msg *operatorv1.SubscribeTasksResponse
	var bid uint32
	var author string

	switch val := val.(type) {
	case *operatorv1.TasksGroupResponse:
		bid = val.GetBid()
		author = val.GetAuthor()
		msg = &operatorv1.SubscribeTasksResponse{
			Type: &operatorv1.SubscribeTasksResponse_Group{
				Group: val,
			},
		}
	case *operatorv1.TasksMessageResponse:
		bid = val.GetBid()
		msg = &operatorv1.SubscribeTasksResponse{
			Type: &operatorv1.SubscribeTasksResponse_Message{
				Message: val,
			},
		}
	case *operatorv1.TasksResponse:
		bid = val.GetBid()
		msg = &operatorv1.SubscribeTasksResponse{
			Type: &operatorv1.SubscribeTasksResponse_Task{
				Task: val,
			},
		}
	case *operatorv1.TasksStatusResponse:
		bid = val.GetBid()
		msg = &operatorv1.SubscribeTasksResponse{
			Type: &operatorv1.SubscribeTasksResponse_TaskStatus{
				TaskStatus: val,
			},
		}
	case *operatorv1.TasksDoneResponse:
		bid = val.GetBid()
		msg = &operatorv1.SubscribeTasksResponse{
			Type: &operatorv1.SubscribeTasksResponse_TaskDone{
				TaskDone: val,
			},
		}
	default:
		return
	}

	// рассылка на подписчиков
	p.pool.Range(func(_ string, v *SubscribeTasks) bool {
		if _, ok := v.bids[bid]; ok {
			// если таск группа невидимая для оператора - пропускаем отправку
			if x, ok := val.(*operatorv1.TasksGroupResponse); ok {
				if !x.Visible && author != v.username {
					return true
				}
			}
			if err := v.ss.Send(msg); err != nil {
				v.err <- err
			}
		}
		return true
	})
}

// Проверка существования оператора (по его username) в мапе.
// Вернет true, если оператор существует в мапе.
func (p *PoolTasks) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *SubscribeTasks) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
		}
		return !f
	})
	return f
}
