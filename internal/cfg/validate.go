package cfg

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	trans ut.Translator
)

// Валидация значения порта
func tagPort(fl validator.FieldLevel) bool {
	if fl.Field().CanInt() {
		// если элемент является int'ом
		n := fl.Field().Int()
		if n > 0 && n < 65535 {
			return true
		}
	}
	return false
}

// Валидация того, что существует только одно из вложенных в структуру полей
func tagOneOfNested(fl validator.FieldLevel) bool {
	// получаем количество вложенных элементов
	n := fl.Field().NumField()
	// обозначаем каунтер вложенных полей (которые defined)
	c := 0
	// итерируем по вложенным структурам
	for i := 0; i < n; i++ {
		// получаем вложенное поле
		f := fl.Field().Field(i)
		if !f.IsZero() {
			// проверяем что поле существует и инкрементируем каунтер
			c++
		}
	}
	// если не найдено единственное поле -> false
	return c == 1
}

// Добавление в валидатор возможности перевода ошибок в человеко-читаемый вид
func translations(v *validator.Validate) error {
	var ok bool
	e := en.New()
	uni := ut.New(e, e)
	trans, ok = uni.GetTranslator("en")
	if !ok {
		return errors.New("could not get translator")
	}
	// "required"
	if err := transRequired(v); err != nil {
		return err
	}
	// "ip"
	if err := transIp(v); err != nil {
		return err
	}
	// "port"
	if err := transPort(v); err != nil {
		return err
	}
	// "one_of_nested"
	if err := transOneOfNested(v); err != nil {
		return err
	}
	return nil
}

// Перевод ошибки тэга "required"
func transRequired(v *validator.Validate) error {
	return v.RegisterTranslation(
		"required",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("required", "{0} must have a value", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("required", fe.StructNamespace())
			if err != nil {
				return fe.(error).Error()
			}
			return t
		},
	)
}

// Перевод ошибки тэга "ip"
func transIp(v *validator.Validate) error {
	return v.RegisterTranslation(
		"ip",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("ip", "{0} must be a valid IP address", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("ip", fe.StructNamespace())
			if err != nil {
				return fe.(error).Error()
			}
			return t
		},
	)
}

// Перевод ошибки тэга "port"
func transPort(v *validator.Validate) error {
	return v.RegisterTranslation(
		"port",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("port", "{0} must be a valid port number (1-65535)", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("port", fe.StructNamespace())
			if err != nil {
				return fe.(error).Error()
			}
			return t
		},
	)
}

// Перевод ошибки тэга "one_of_nested"
func transOneOfNested(v *validator.Validate) error {
	return v.RegisterTranslation(
		"one_of_nested",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("one_of_nested", "Only one value must be present in {0}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("one_of_nested", fe.StructNamespace())
			if err != nil {
				return fe.(error).Error()
			}
			return t
		},
	)
}
