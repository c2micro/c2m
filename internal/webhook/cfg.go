package webhook

import (
	"context"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Config struct {
	Telegram ConfigTelegram `json:"telegram" validate:"omitempty"`
	Vkteams  ConfigVkteams  `json:"vkteams" validate:"omitempty"`
}

// Инициализация вебхуков. Если появляются ошибки в процессе -> просто игнорируем и пишем в лог
func (cfg Config) Init(ctx context.Context) {
	lg := zctx.From(ctx).Named("webhook")

	// добавление Telegram
	telegram, err := newTelegram(ctx, cfg.Telegram)
	if err != nil {
		lg.Error("unable initialize telegram webhook", zap.Error(err))
	}
	if telegram != nil {
		Webhook.telegram = telegram
		lg.Debug("registered telegram webhook")
	}

	// Добавление Vkteams
	vkteams, err := newVkteams(ctx, cfg.Vkteams)
	if err != nil {
		lg.Error("unable initializa vkteams webhook", zap.Error(err))
	}
	if vkteams != nil {
		Webhook.vkteams = vkteams
		lg.Debug("registered vkteams webhook")
	}
}
