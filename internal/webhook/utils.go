package webhook

import (
	"bytes"
	"html/template"
	"strings"
)

var shouldBeEscaped = "_*[]()~`>#+-=|{}.!"

// Подстановка данных из структуры в шаблон для получения итоговой строки
func executeTmpl(t *template.Template, d *TemplateData) (string, error) {
	buf := &bytes.Buffer{}
	err := t.Execute(buf, d)
	return buf.String(), err
}

// Санитизация пользовательских данных для корректного парсинга MarkdownV2
func escapeMarkdown(s string) string {
	var result []rune
	var escaped bool
	for _, r := range s {
		if r == '\\' {
			escaped = !escaped
			result = append(result, r)
			continue
		}
		if strings.ContainsRune(shouldBeEscaped, r) && !escaped {
			result = append(result, '\\')
		}
		escaped = false
		result = append(result, r)
	}
	return string(result)
}
