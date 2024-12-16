package cfg

import (
	"context"
)

// Тип-заглушка для ключа в контексте
type configCtxKey struct{}

// Добавление в контекст конфиг
func SetConfigCtx(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, configCtxKey{}, c)
}

// Получение конфига из контекста
func GetConfigCtx(ctx context.Context) Config {
	return ctx.Value(configCtxKey{}).(Config)
}
