package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/c2micro/c2mshr/defaults"
	"github.com/c2micro/c2m/internal/constants"
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
			MaxLen(defaults.CredentialUsernameMaxLength).
			Optional().
			Comment("username of credential"),
		field.String("secret").
			MaxLen(defaults.CredentialSecretMaxLength).
			Optional().
			Comment("password/hash/secret"),
		field.String("realm").
			MaxLen(defaults.CredentialRealmMaxLength).
			Optional().
			Comment("realm of host (or zone where credentials are valid)"),
		field.String("host").
			MaxLen(defaults.CredentialHostMaxLength).
			Optional().
			Comment("host from which creds has been extracted"),
		field.String("note").
			MaxLen(defaults.CredentialNoteMaxLength).
			Optional().
			Comment("note of credential"),
		field.Uint32("color").
			Default(constants.DefaultColor).
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
