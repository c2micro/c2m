package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"

	"github.com/lrita/cmap"
)

type SubscribeChat struct {
	username   string
	disconnect chan struct{}
	data       chan any
	err        chan error
	ss         operatorv1.OperatorService_SubscribeChatServer
}

type PoolChat struct {
	pool cmap.Map[string, *SubscribeChat]
}

// Получение ошибки, возникшей при обработке сессии оператора.
func (s *SubscribeChat) Error() chan error {
	return s.err
}

// Получение нотификации о небходимости отключения оператора от подписки.
func (s *SubscribeChat) IsDisconnect() chan struct{} {
	return s.disconnect
}

// Добавление сессии оператора в мапу.
func (p *PoolChat) Add(c, u string, s operatorv1.OperatorService_SubscribeChatServer) {
	p.pool.Store(c, &SubscribeChat{
		username:   u,
		disconnect: make(chan struct{}, 1),
		data:       make(chan any, 1),
		err:        make(chan error),
		ss:         s,
	})
}

// Нотификация о необходимости удаления сессии оператора из мапы.
func (p *PoolChat) Disconnect(c string) {
	v, ok := p.pool.Load(c)
	if ok {
		v.disconnect <- struct{}{}
	}
}

// Удаление сессии оператора из мапы.
func (p *PoolChat) Remove(c string) {
	p.pool.Delete(c)
}

// Получение сессии оператора из мапы.
func (p *PoolChat) Get(c string) *SubscribeChat {
	o, _ := p.pool.Load(c)
	return o
}

// Отправка сообщения всем операторам, подписанным на топик с чатом.
func (p *PoolChat) Send(val any) {
	var msg *operatorv1.SubscribeChatResponse

	switch val := val.(type) {
	case *operatorv1.ChatResponse:
		msg = &operatorv1.SubscribeChatResponse{
			Response: &operatorv1.SubscribeChatResponse_Message{
				Message: val,
			},
		}
	case *operatorv1.ChatMessagesResponse:
		msg = &operatorv1.SubscribeChatResponse{
			Response: &operatorv1.SubscribeChatResponse_Messages{
				Messages: val,
			},
		}
	default:
		return
	}

	// рассылка на подписчиков
	p.pool.Range(func(_ string, v *SubscribeChat) bool {
		if err := v.ss.Send(msg); err != nil {
			v.err <- err
		}
		return true
	})
}

// Проверка существования оператора (по его username) в мапе.
// Вернет true, если оператор существует в мапе.
func (p *PoolChat) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *SubscribeChat) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
		}
		return !f
	})
	return f
}
