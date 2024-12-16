package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/lrita/cmap"
)

type SubscribeCredentials struct {
	username   string
	disconnect chan struct{}
	data       chan any
	err        chan error
	ss         operatorv1.OperatorService_SubscribeCredentialsServer
}

type PoolCredentials struct {
	pool cmap.Map[string, *SubscribeCredentials]
}

// Получение ошибки, возникшей при обработке сессии оператора.
func (s *SubscribeCredentials) Error() chan error {
	return s.err
}

// Получение нотификации о небходимости отключения оператора от подписки.
func (s *SubscribeCredentials) IsDisconnect() chan struct{} {
	return s.disconnect
}

// Добавление сессии оператора в мапу.
func (p *PoolCredentials) Add(c, u string, s operatorv1.OperatorService_SubscribeCredentialsServer) {
	p.pool.Store(c, &SubscribeCredentials{
		username:   u,
		disconnect: make(chan struct{}, 1),
		data:       make(chan any, 1),
		err:        make(chan error),
		ss:         s,
	})
}

// Нотификация о необходимости удаления сессии оператора из мапы.
func (p *PoolCredentials) Disconnect(c string) {
	v, ok := p.pool.Load(c)
	if ok {
		v.disconnect <- struct{}{}
	}
}

// Удаление сессии оператора из мапы.
func (p *PoolCredentials) Remove(c string) {
	p.pool.Delete(c)
}

// Получение сессии оператора из мапы.
func (p *PoolCredentials) Get(c string) *SubscribeCredentials {
	o, _ := p.pool.Load(c)
	return o
}

// Отправка сообщения всем операторам, подписанным на топик с кредылами.
func (p *PoolCredentials) Send(val any) {
	var msg *operatorv1.SubscribeCredentialsResponse

	switch val := val.(type) {
	case *operatorv1.CredentialResponse:
		msg = &operatorv1.SubscribeCredentialsResponse{
			Response: &operatorv1.SubscribeCredentialsResponse_Credential{
				Credential: val,
			},
		}
	case *operatorv1.CredentialColorResponse:
		msg = &operatorv1.SubscribeCredentialsResponse{
			Response: &operatorv1.SubscribeCredentialsResponse_Color{
				Color: val,
			},
		}
	case *operatorv1.CredentialNoteResponse:
		msg = &operatorv1.SubscribeCredentialsResponse{
			Response: &operatorv1.SubscribeCredentialsResponse_Note{
				Note: val,
			},
		}
	case *operatorv1.CredentialsResponse:
		msg = &operatorv1.SubscribeCredentialsResponse{
			Response: &operatorv1.SubscribeCredentialsResponse_Credentials{
				Credentials: val,
			},
		}
	default:
		return
	}

	// рассылка на подписчиков
	p.pool.Range(func(_ string, v *SubscribeCredentials) bool {
		if err := v.ss.Send(msg); err != nil {
			v.err <- err
		}
		return true
	})
}

// Проверка существования оператора (по его username) в мапе.
// Вернет true, если оператор существует в мапе.
func (p *PoolCredentials) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *SubscribeCredentials) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
		}
		return !f
	})
	return f
}
