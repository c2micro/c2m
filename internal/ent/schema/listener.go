package schema

import (
	"fmt"
	"net"
	"time"

	"github.com/c2micro/c2mshr/defaults"
	"github.com/c2micro/c2m/internal/constants"
	"github.com/c2micro/c2m/internal/types"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Listener объявление схемы Listener
type Listener struct {
	ent.Schema
}

// Annotations аннотации для Listener
func (Listener) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "listener",
		},
	}
}

// Fields поля для модели Listener
func (Listener) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			MinLen(32).
			MaxLen(32).
			Unique().
			Optional().
			Comment("authentication token of listener"),
		field.String("name").
			MaxLen(defaults.ListenerNameMaxLength).
			Optional().
			Comment("freehand name of listener"),
		field.String("ip").
			GoType(types.Inet{}).
			SchemaType(map[string]string{
				dialect.Postgres: "inet",
			}).
			Validate(func(s string) error {
				if net.ParseIP(s) == nil {
					return fmt.Errorf("invalid value of ip %q", s)
				}
				return nil
			}).
			Optional().
			Comment("bind ip address of listener"),
		field.Uint16("port").
			Min(1).
			Optional().
			Comment("bind port of listener"),
		field.Uint32("color").
			Default(constants.DefaultColor).
			Comment("color of entity"),
		field.String("note").
			MaxLen(defaults.ListenerNoteMaxLength).
			Optional().
			Comment("note of listener"),
		field.Time("last").
			Default(func() time.Time {
				return time.Time{}
			}).
			Comment("last activity of listener"),
	}
}

// Edges связи для Listener
func (Listener) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("beacon", Beacon.Type),
	}
}

// Mixin mixin для Listener
func (Listener) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
