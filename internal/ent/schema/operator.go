package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/c2micro/c2m/internal/constants"
)

// Operator объявление схемы Operator
type Operator struct {
	ent.Schema
}

// Annotations аннотации для Operator
func (Operator) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "operator",
		},
	}
}

// Fields поля для модели Operator
func (Operator) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			MaxLen(255).
			Comment("username of operator"),
		field.String("token").
			MinLen(32).
			MaxLen(32).
			Optional().
			Comment("access token for operator"),
		field.Uint32("color").
			Default(constants.DefaultColor).
			Comment("color of entity"),
		field.Time("last").
			Default(func() time.Time {
				return time.Time{}
			}).
			Comment("last activity of operator"),
	}
}

// Edges связи для Operator
func (Operator) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("chat", Chat.Type),
		edge.To("group", Group.Type),
	}
}

// Mixin mixin для Operator
func (Operator) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
