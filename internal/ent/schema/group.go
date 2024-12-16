package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/c2micro/c2msrv/internal/defaults"
)

// Group объявление схемы Group
type Group struct {
	ent.Schema
}

// Annotations аннотации для Group
func (Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "group",
		},
	}
}

// Fields поля для модели Group
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.Int("bid").
			Comment("beacon ID"),
		field.String("cmd").
			MinLen(defaults.GroupMinCmdLen).
			MaxLen(defaults.GroupMaxCmdLen).
			Comment("command with arguments"),
		field.Bool("visible").
			Comment("is group visible for other operators"),
		field.Int("author").
			Comment("author of group"),
		field.Time("created_at").
			Default(time.Now).
			Comment("when group created"),
		field.Time("closed_at").
			Optional().
			Comment("when group closed"),
	}
}

// Edges связи для Group
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("beacon", Beacon.Type).
			Ref("group").
			Field("bid").
			Unique().
			Required(),
		edge.From("operator", Operator.Type).
			Ref("group").
			Field("author").
			Unique().
			Required(),
		edge.To("message", Message.Type),
		edge.To("task", Task.Type),
	}
}
