package webhook

import (
	"context"
	"html/template"

	"github.com/go-faster/sdk/zctx"
	bot "github.com/mail-ru-im/bot-golang"
	"go.uber.org/zap"
)

type ConfigVkteams struct {
	Enabled bool   `json:"enabled" default:"false"`
	Token   string `json:"token" validate:"required_if=Enabled true"`
	ChatId  string `json:"chat_id" validate:"required_if=Enabled true"`
	Text    string `json:"text" validate:"required_if=Enabled true"`
	Api     string `json:"api" validate:"required_if=Enabled true"`
}

type Vkteams struct {
	bot      *bot.Bot
	ctx      context.Context
	chatId   string
	template *template.Template
}

// Создание объекта Vkteams для возможности отправки сообщений через бота
func newVkteams(ctx context.Context, cfg ConfigVkteams) (*Vkteams, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	// парсинг шаблона
	tmpl, err := template.New("vkteams").Parse(cfg.Text)
	if err != nil {
		return nil, err
	}
	// создание бота
	if b, err := bot.NewBot(cfg.Token, bot.BotApiURL(cfg.Api)); err != nil {
		return nil, err
	} else {
		return &Vkteams{
			bot:      b,
			ctx:      ctx,
			chatId:   cfg.ChatId,
			template: tmpl,
		}, nil
	}
}

// Отправка сообщения в Vkteams
func (v *Vkteams) Send(data *TemplateData) {
	lg := zctx.From(v.ctx).Named("vkteams")

	// санитизация данных для MarkdownV2
	data.escape()
	// заполнение шаблона
	text, err := executeTmpl(v.template, data)
	if err != nil {
		lg.Error("unable compile message", zap.Error(err))
	}
	// отправка сообщения
	if err := v.bot.SendMessage(&bot.Message{
		Chat:        bot.Chat{ID: v.chatId},
		Text:        text,
		ParseMode:   bot.ParseModeMarkdownV2,
		ContentType: bot.Text,
	}); err != nil {
		lg.Error("unable send webhook via vkteams", zap.Error(err))
	}
}
