package pools

import (
	"strings"

	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"

	"github.com/lrita/cmap"
)

type subscribeHello struct {
	username string
	ss       operatorv1.OperatorService_HelloServer
}

type PoolHello struct {
	pool cmap.Map[string, *subscribeHello]
}

// Добавление сессии оператора в хранилище.
// Данное действие соотносимо с авторизацией.
func (p *PoolHello) Add(c, u string, s operatorv1.OperatorService_HelloServer) {
	p.pool.Store(c, &subscribeHello{
		username: u,
		ss:       s,
	})
}

// Удаление сессии оператора из хранилища.
// Данное действие соотносимо с обнулением сессии.
func (p *PoolHello) Remove(c string) {
	p.pool.Delete(c)
}

// Проверка наличия оператора в Hello мапе.
// Вернет true, если оператор залогинен.
func (p *PoolHello) Exists(u string) bool {
	f := false
	p.pool.Range(func(c string, v *subscribeHello) bool {
		if strings.Compare(v.username, u) == 0 {
			f = true
			return false
		}
		return true
	})
	return f
}

// Валидация сессионной куки оператора.
// Вернет true, если кука соответствует оператору.
func (p *PoolHello) Validate(u, c string) bool {
	if v, ok := p.pool.Load(c); ok {
		if strings.Compare(v.username, u) == 0 {
			return true
		}
	}
	return false
}
