package webhook

import (
	"context"
	"html/template"

	"github.com/go-faster/sdk/zctx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type ConfigTelegram struct {
	Enabled bool   `json:"enabled" default:"false"`
	Token   string `json:"token" validate:"required_if=Enabled true"`
	ChatId  string `json:"chat_id" validate:"required_if=Enabled true"`
	Text    string `json:"text" validate:"required_if=Enabled true"`
}

type Telegram struct {
	bot      *bot.Bot
	ctx      context.Context
	chatId   string
	template *template.Template
}

// Создание объекта Telegram для возможности отправки сообщений через бота
func newTelegram(ctx context.Context, cfg ConfigTelegram) (*Telegram, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	// парсинг шаблона
	tmpl, err := template.New("telegram").Parse(cfg.Text)
	if err != nil {
		return nil, err
	}
	// создание бота
	if b, err := bot.New(cfg.Token); err != nil {
		return nil, err
	} else {
		return &Telegram{
			bot:      b,
			ctx:      ctx,
			chatId:   cfg.ChatId,
			template: tmpl,
		}, nil
	}
}

// Отправка сообщения в Telegram
func (t *Telegram) Send(data *TemplateData) {
	lg := zctx.From(t.ctx).Named("telegram")

	// санитизация данных для MarkdownV2
	data.escape()
	// заполнение шаблона
	text, err := executeTmpl(t.template, data)
	if err != nil {
		lg.Error("unable compile message", zap.Error(err))
	}
	// отправка сообщения
	if _, err := t.bot.SendMessage(t.ctx, &bot.SendMessageParams{
		ChatID:    t.chatId,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	}); err != nil {
		lg.Error("unable send webhook via telegram", zap.Error(err))
	}
}
