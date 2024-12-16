package schema

import (
	"context"
	"fmt"
	"time"

	gen "github.com/c2micro/c2msrv/internal/ent"
	"github.com/c2micro/c2msrv/internal/ent/hook"
	"github.com/c2micro/c2msrv/internal/ent/intercept"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// TimeMixin структура для хранения временных меток
type TimeMixin struct {
	mixin.Schema
}

// Fields поля для модели TimeMixin
func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Comment("Time when entity was created").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			Comment("Time when entity was updated").
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Comment("Time when entity was soft-deleted"),
	}
}

// Interceptors для модели TimeMixin
func (s TimeMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(c context.Context, q intercept.Query) error {
			if i, _ := c.Value(softDeleteKey{}).(bool); i {
				// i == true означает включение в выборку soft-deleted записей
				return nil
			}
			s.P(q)
			return nil
		}),
	}
}

// Hooks хуки для модели TimeMixin
func (s TimeMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					if del, _ := ctx.Value(softDeleteKey{}).(bool); del {
						// del == true означает полное удаление записи
						return next.Mutate(ctx, m)
					}
					mx, ok := m.(interface {
						SetOp(ent.Op)
						Client() *gen.Client
						SetDeletedAt(time.Time)
						WhereP(...func(*sql.Selector))
					})
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type %T", m)
					}
					s.P(mx)
					mx.SetOp(ent.OpUpdate)
					mx.SetDeletedAt(time.Now())
					return mx.Client().Mutate(ctx, m)
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}

// P добавление предиката для уровня хранения (используется в запросах и мутациях)
func (s TimeMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		// [2] - id поля "deleted_at" в массиве []ent.Field{}
		sql.FieldIsNull(s.Fields()[2].Descriptor().Name),
	)
}

// softDeleteKey структура-заглушку для ключа в контексте
type softDeleteKey struct{}

// SkipSoftDelete контекст, который позволит удалить полностью запись
func SkipSoftDelete(c context.Context) context.Context {
	return context.WithValue(c, softDeleteKey{}, true)
}

// WithSoftDeleted контекст, который позволит добавить в выборку soft-deleted записи
func WithSoftDeleted(c context.Context) context.Context {
	return context.WithValue(c, softDeleteKey{}, true)
}
