package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"

	"github.com/lrita/cmap"
)

type SubscribeBeacons struct {
	username   string
	disconnect chan struct{}
	data       chan any
	err        chan error
	ss         operatorv1.OperatorService_SubscribeBeaconsServer
}

type PoolBeacons struct {
	pool cmap.Map[string, *SubscribeBeacons]
}

// Получение ошибки, возникшей при обработке сессии оператора.
func (s *SubscribeBeacons) Error() chan error {
	return s.err
}

// Получение нотификации о небходимости отключения оператора от подписки.
func (s *SubscribeBeacons) IsDisconnect() chan struct{} {
	return s.disconnect
}

// Добавление сессии оператора в мапу.
func (p *PoolBeacons) Add(c, u string, s operatorv1.OperatorService_SubscribeBeaconsServer) {
	p.pool.Store(c, &SubscribeBeacons{
		username:   u,
		disconnect: make(chan struct{}, 1),
		data:       make(chan any, 1),
		err:        make(chan error),
		ss:         s,
	})
}

// Нотификация о необходимости удаления сессии оператора из мапы.
func (p *PoolBeacons) Disconnect(c string) {
	v, ok := p.pool.Load(c)
	if ok {
		v.disconnect <- struct{}{}
	}
}

// Удаление сессии оператора из мапы.
func (p *PoolBeacons) Remove(c string) {
	p.pool.Delete(c)
}

// Получение сессии оператора из мапы.
func (p *PoolBeacons) Get(c string) *SubscribeBeacons {
	b, _ := p.pool.Load(c)
	return b
}

// Отправка сообщения всем операторам, подписанным на топик с биконами.
func (p *PoolBeacons) Send(val any) {
	var msg *operatorv1.SubscribeBeaconsResponse

	switch val := val.(type) {
	case *operatorv1.BeaconResponse:
		msg = &operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Beacon{
				Beacon: val,
			},
		}
	case *operatorv1.BeaconColorResponse:
		msg = &operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Color{
				Color: val,
			},
		}
	case *operatorv1.BeaconNoteResponse:
		msg = &operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Note{
				Note: val,
			},
		}
	case *operatorv1.BeaconLastResponse:
		msg = &operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Last{
				Last: val,
			},
		}
	case *operatorv1.BeaconSleepResponse:
		msg = &operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Sleep{
				Sleep: val,
			},
		}
	case *operatorv1.BeaconsResponse:
		msg = &operatorv1.SubscribeBeaconsResponse{
			Response: &operatorv1.SubscribeBeaconsResponse_Beacons{
				Beacons: val,
			},
		}
	default:
		return
	}

	// рассылка на подписчиков
	p.pool.Range(func(_ string, v *SubscribeBeacons) bool {
		if err := v.ss.Send(msg); err != nil {
			v.err <- err
		}
		return true
	})
}

// Проверка существования оператора (по его username) в мапе.
// Вернет true, если оператор существует в мапе.
func (p *PoolBeacons) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *SubscribeBeacons) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
		}
		return !f
	})
	return f
}
