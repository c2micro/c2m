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

// Task объявление схемы Task
type Task struct {
	ent.Schema
}

// Annotations аннотации для Task
func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "task",
		},
	}
}

// Fields поля для модели Task
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.Int("gid").
			Comment("id of task group"),
		field.Int("bid").
			Comment("id of beacon"),
		field.Time("created_at").
			Default(time.Now).
			Comment("time when task created"),
		field.Time("pushed_at").
			Optional().
			Comment("time when task pushed to the beacon"),
		field.Time("done_at").
			Optional().
			Comment("time when task results received"),
		field.Enum("status").
			GoType(defaults.TaskStatus(0)).
			Comment("status of task"),
		field.Enum("cap").
			GoType(defaults.Capability(0)).
			Comment("capability to execute"),
		field.Int("args").
			Comment("capability arguments"),
		field.Int("output").
			Optional().
			Comment("task output"),
		field.Bool("output_big").
			Optional().
			Comment("is output bigger than constant value"),
	}
}

// Edges связи для Task
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("task").
			Field("gid").
			Unique().
			Required(),
		edge.From("beacon", Beacon.Type).
			Ref("task").
			Field("bid").
			Unique().
			Required(),
		edge.From("blobber_args", Blobber.Type).
			Ref("task_args").
			Field("args").
			Unique().
			Required(),
		edge.From("blobber_output", Blobber.Type).
			Ref("task_output").
			Field("output").
			Unique(),
	}
}
