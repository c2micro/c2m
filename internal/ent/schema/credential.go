package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/c2micro/c2msrv/internal/defaults"
)

// Credential объявление схемы Credential
type Credential struct {
	ent.Schema
}

// Annotations аннотации для Credential
func (Credential) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "credential",
		},
	}
}

// Fields поля для модели Credential
func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			MaxLen(defaults.CredentialMaxLenUsername).
			Optional().
			Comment("username of credential"),
		field.String("password").
			MaxLen(defaults.CredentialMaxLenPassword).
			Optional().
			Comment("password/hash/secret"),
		field.String("realm").
			MaxLen(defaults.CredentialMaxLenRealm).
			Optional().
			Comment("realm of host (or zone where credentials are valid)"),
		field.String("host").
			MaxLen(defaults.CredentialMaxLenHost).
			Optional().
			Comment("host from which creds has been extracted"),
		field.String("note").
			MaxLen(defaults.CredentialMaxLenNote).
			Optional().
			Comment("note of credential"),
		field.Uint32("color").
			Default(defaults.DefaultColor).
			Comment("color of entity"),
	}
}

// Edges связи для Credential
func (Credential) Edges() []ent.Edge {
	return nil
}

// Mixin mixin для Beacon
func (Credential) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
