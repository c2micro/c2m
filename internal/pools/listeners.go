package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"

	"github.com/lrita/cmap"
)

type SubscribeListeners struct {
	username   string
	disconnect chan struct{}
	data       chan any
	err        chan error
	ss         operatorv1.OperatorService_SubscribeListenersServer
}

type PoolListeners struct {
	pool cmap.Map[string, *SubscribeListeners]
}

// Получение ошибки, возникшей при обработке сессии оператора.
func (s *SubscribeListeners) Error() chan error {
	return s.err
}

// Получение нотификации о небходимости отключения оператора от подписки.
func (s *SubscribeListeners) IsDisconnect() chan struct{} {
	return s.disconnect
}

// Добавление сессии оператора в мапу.
func (p *PoolListeners) Add(c, u string, s operatorv1.OperatorService_SubscribeListenersServer) {
	p.pool.Store(c, &SubscribeListeners{
		username:   u,
		disconnect: make(chan struct{}, 1),
		data:       make(chan any, 1),
		err:        make(chan error),
		ss:         s,
	})
}

// Нотификация о необходимости удаления сессии оператора из мапы.
func (p *PoolListeners) Disconnect(c string) {
	v, ok := p.pool.Load(c)
	if ok {
		v.disconnect <- struct{}{}
	}
}

// Удаление сессии оператора из мапы.
func (p *PoolListeners) Remove(c string) {
	p.pool.Delete(c)
}

// Получение сессии оператора из мапы.
func (p *PoolListeners) Get(c string) *SubscribeListeners {
	l, _ := p.pool.Load(c)
	return l
}

// Отправка сообщения всем операторам, подписанным на топик с листенерами.
func (p *PoolListeners) Send(val any) {
	var msg *operatorv1.SubscribeListenersResponse

	switch val := val.(type) {
	case *operatorv1.ListenerResponse:
		msg = &operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Listener{
				Listener: val,
			},
		}
	case *operatorv1.ListenerColorResponse:
		msg = &operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Color{
				Color: val,
			},
		}
	case *operatorv1.ListenerNoteResponse:
		msg = &operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Note{
				Note: val,
			},
		}
	case *operatorv1.ListenerInfoResponse:
		msg = &operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Info{
				Info: val,
			},
		}
	case *operatorv1.ListenerLastResponse:
		msg = &operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Last{
				Last: val,
			},
		}
	case *operatorv1.ListenersResponse:
		msg = &operatorv1.SubscribeListenersResponse{
			Response: &operatorv1.SubscribeListenersResponse_Listeners{
				Listeners: val,
			},
		}
	default:
		return
	}

	// рассылка на подписчиков
	p.pool.Range(func(_ string, v *SubscribeListeners) bool {
		if err := v.ss.Send(msg); err != nil {
			v.err <- err
		}
		return true
	})
}

// Проверка существования оператора (по его username) в мапе.
// Вернет true, если оператор существует в мапе.
func (p *PoolListeners) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *SubscribeListeners) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
		}
		return !f
	})
	return f
}
