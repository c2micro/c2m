package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Blobber holds the schema definition for the Blobber entity.
type Blobber struct {
	ent.Schema
}

// Annotations of Blobber.
func (Blobber) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "blobber",
		},
	}
}

// Fields поля для модели Blobber
func (Blobber) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("hash").
			Unique().
			Comment("non-cryptographic hash of blob"),
		field.Bytes("blob").
			Comment("blob to store"),
		field.Int("size").
			Comment("real size of blob"),
	}
}

// Edges of the Blobber.
func (Blobber) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("task_args", Task.Type),
		edge.To("task_output", Task.Type),
	}
}

// Mixin шаринг модели для Blobber
func (Blobber) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
