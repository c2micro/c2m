package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/c2micro/c2mshr/defaults"
)

// Chat объявление схемы Chat
type Chat struct {
	ent.Schema
}

// Annotations аннотации для Chat
func (Chat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "chat",
		},
	}
}

// Fields поля для модели Chat
func (Chat) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("when message created"),
		field.Int("from").
			Optional().
			Comment("creator of message"),
		field.String("message").
			MinLen(defaults.ChatMessageMinLength).
			MaxLen(defaults.ChatMessageMaxLength).
			Comment("message itself"),
	}
}

// Edges связи для Chat
func (Chat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("operator", Operator.Type).
			Ref("chat").
			Field("from").
			Unique(),
	}
}
