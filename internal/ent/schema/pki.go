package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Pki объявление схемы Pki
type Pki struct {
	ent.Schema
}

// Annotations аннотации для Pki
func (Pki) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "pki",
		},
	}
}

// Fields поля для модели Pki
func (Pki) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").
			Values("ca", "listener", "operator", "management").
			Comment("type of certificate blob (ca, listener, operator)"),
		field.Bytes("key").
			Comment("certificate key"),
		field.Bytes("cert").
			Comment("certificate chain"),
	}
}

// Edges связи для Pki
func (Pki) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Mixin mixin для Operator
func (Pki) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
