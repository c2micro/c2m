package webhook

// Объект для хранения и обработки перенаправления сообщений в разные коннекторы
var Webhook WebhookConnectors = WebhookConnectors{}

type WebhookConnectors struct {
	telegram *Telegram
	vkteams  *Vkteams
}

// Отправка сообщения во все коннекторы
func (w WebhookConnectors) Send(data *TemplateData) {
	if w.telegram != nil {
		go w.telegram.Send(data)
	}
	if w.vkteams != nil {
		go w.vkteams.Send(data)
	}
}

type TemplateData struct {
	Bid        int
	Username   string
	Hostname   string
	Domain     string
	ExternalIP string
	InternalIP string
	Privileged bool
}

// Санитизация структруы
func (td *TemplateData) escape() {
	td.Username = escapeMarkdown(td.Username)
	td.Hostname = escapeMarkdown(td.Hostname)
	td.Domain = escapeMarkdown(td.Domain)
	td.ExternalIP = escapeMarkdown(td.ExternalIP)
	td.InternalIP = escapeMarkdown(td.InternalIP)
}
