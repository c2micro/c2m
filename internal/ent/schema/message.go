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

// Message объявление схемы Message
type Message struct {
	ent.Schema
}

// Annotations аннотации для Message
func (Message) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "message",
		},
	}
}

// Fields поля для модели Message
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.Int("gid").
			Comment("id of task group"),
		field.Enum("type").
			GoType(defaults.TaskMessage(0)).
			Comment("type of message"),
		field.String("message").
			MaxLen(defaults.TaskGroupMessageMaxLength).
			Comment("message itself"),
		field.Time("created_at").
			Default(time.Now).
			Comment("when message created"),
	}
}

// Edges связи для Message
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("message").
			Field("gid").
			Unique().
			Required(),
	}
}
