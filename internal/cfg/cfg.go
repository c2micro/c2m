package cfg

import (
	"context"
	"errors"
	"os"

	"github.com/c2micro/c2msrv/internal/db"
	"github.com/c2micro/c2msrv/internal/listener"
	"github.com/c2micro/c2msrv/internal/management"
	"github.com/c2micro/c2msrv/internal/operator"
	"github.com/c2micro/c2msrv/internal/pki"
	"github.com/c2micro/c2msrv/internal/webhook"

	"github.com/go-faster/sdk/zctx"
	"github.com/go-playground/validator/v10"
	"sigs.k8s.io/yaml"
)

// Структура конфигурационного файла
type Config struct {
	Debug      bool                `json:"debug" validate:"omitempty"`
	Listener   listener.ConfigV1   `json:"listener" validate:"required"`
	Operator   operator.ConfigV1   `json:"operator" validate:"required"`
	Management management.ConfigV1 `json:"management" validate:"required"`
	Db         db.ConfigV1         `json:"db" validate:"required,one_of_nested"`
	Pki        pki.ConfigV1        `json:"pki"`
	Webhook    webhook.Config      `json:"webhook"`
}

// Вычитывание конфиг файла и маппинг полей
func Read(path string) (Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err = yaml.UnmarshalStrict(f, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Валидация полей структуры из конфига
func (c Config) Validate(ctx context.Context) error {
	lg := zctx.From(ctx).Named("cfg-validator")

	v := validator.New(validator.WithRequiredStructEnabled())

	// регистрируем кастомный валидатор "one_of_nested"
	if err := v.RegisterValidation("one_of_nested", tagOneOfNested); err != nil {
		return err
	}
	// регистрируем кастомный валидатор "port"
	if err := v.RegisterValidation("port", tagPort); err != nil {
		return err
	}
	// регистрируем перевод ошибок
	if err := translations(v); err != nil {
		return err
	}
	// валидируем конфиг
	if err := v.Struct(c); err != nil {
		var es validator.ValidationErrors
		if errors.As(err, &es) {
			for _, e := range es {
				lg.Error(e.Translate(trans))
			}
		}
		return errors.New("bunch of validation errors")
	}
	return nil
}
