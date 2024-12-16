package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"

	"github.com/lrita/cmap"
)

type SubscribeOperators struct {
	username   string
	disconnect chan struct{}
	data       chan any
	err        chan error
	ss         operatorv1.OperatorService_SubscribeOperatorsServer
}

type PoolOperators struct {
	pool cmap.Map[string, *SubscribeOperators]
}

// Получение ошибки, возникшей при обработке сессии оператора.
func (s *SubscribeOperators) Error() chan error {
	return s.err
}

// Получение нотификации о небходимости отключения оператора от подписки.
func (s *SubscribeOperators) IsDisconnect() chan struct{} {
	return s.disconnect
}

// Добавление сессии оператора в мапу.
func (p *PoolOperators) Add(c, u string, s operatorv1.OperatorService_SubscribeOperatorsServer) {
	p.pool.Store(c, &SubscribeOperators{
		username:   u,
		disconnect: make(chan struct{}, 1),
		data:       make(chan any, 1),
		err:        make(chan error),
		ss:         s,
	})
}

// Нотификация о необходимости удаления сессии оператора из мапы.
func (p *PoolOperators) Disconnect(c string) {
	v, ok := p.pool.Load(c)
	if ok {
		v.disconnect <- struct{}{}
	}
}

// Удаление сессии оператора из мапы.
func (p *PoolOperators) Remove(c string) {
	p.pool.Delete(c)
}

// Получение сессии оператора из мапы.
func (p *PoolOperators) Get(c string) *SubscribeOperators {
	o, _ := p.pool.Load(c)
	return o
}

// Отправка сообщения всем операторам, подписанным на топик с операторами.
func (p *PoolOperators) Send(val any) {
	var msg *operatorv1.SubscribeOperatorsResponse

	switch val := val.(type) {
	case *operatorv1.OperatorLastResponse:
		msg = &operatorv1.SubscribeOperatorsResponse{
			Response: &operatorv1.SubscribeOperatorsResponse_Last{
				Last: val,
			},
		}
	case *operatorv1.OperatorResponse:
		msg = &operatorv1.SubscribeOperatorsResponse{
			Response: &operatorv1.SubscribeOperatorsResponse_Operator{
				Operator: val,
			},
		}
	case *operatorv1.OperatorColorResponse:
		msg = &operatorv1.SubscribeOperatorsResponse{
			Response: &operatorv1.SubscribeOperatorsResponse_Color{
				Color: val,
			},
		}
	case *operatorv1.OperatorsResponse:
		msg = &operatorv1.SubscribeOperatorsResponse{
			Response: &operatorv1.SubscribeOperatorsResponse_Operators{
				Operators: val,
			},
		}
	default:
		return
	}

	// рассылка на подписчиков
	p.pool.Range(func(_ string, v *SubscribeOperators) bool {
		if err := v.ss.Send(msg); err != nil {
			v.err <- err
		}
		return true
	})
}

// Проверка существования оператора (по его username) в мапе.
// Вернет true, если оператор существует в мапе.
func (p *PoolOperators) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *SubscribeOperators) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
		}
		return !f
	})
	return f
}
