package pools

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Глобальный пул для работы с подписками операторов на топики
var Pool pool

// Локальная структура для хранения всех типов пулов
type pool struct {
	Hello       PoolHello
	Chat        PoolChat
	Listeners   PoolListeners
	Beacons     PoolBeacons
	Operators   PoolOperators
	Credentials PoolCredentials
	Tasks       PoolTasks
}

// Отключение оператора от всех топиков
func (p *pool) DisconnectAll(ctx context.Context, c string) {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		p.Chat.Disconnect(c)
		return nil
	})
	g.Go(func() error {
		p.Listeners.Disconnect(c)
		return nil
	})
	g.Go(func() error {
		p.Beacons.Disconnect(c)
		return nil
	})
	g.Go(func() error {
		p.Operators.Disconnect(c)
		return nil
	})
	g.Go(func() error {
		p.Credentials.Disconnect(c)
		return nil
	})
	g.Go(func() error {
		p.Tasks.Disconnect(c)
		return nil
	})
	_ = g.Wait()
}
