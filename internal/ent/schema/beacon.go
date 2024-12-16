package schema

import (
	"fmt"
	"net"
	"time"

	c "github.com/c2micro/c2mshr/defaults"
	"github.com/c2micro/c2msrv/internal/defaults"
	"github.com/c2micro/c2msrv/internal/types"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Beacon объявление схемы Beacon
type Beacon struct {
	ent.Schema
}

// Annotations аннотации для Beacon
func (Beacon) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			// название таблицы
			Table: "beacon",
		},
	}
}

// Fields поля для модели Beacon
func (Beacon) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("bid").
			Unique().
			Comment("beacon ID"),
		field.Int("listener_id").
			Comment("linked listener ID"),
		field.String("ext_ip").
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
			Comment("external IP address of beacon"),
		field.String("int_ip").
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
			Comment("internal IP address of beacon"),
		field.Enum("os").
			GoType(c.BeaconOS(0)).
			Comment("type of operating system"),
		field.String("os_meta").
			MaxLen(defaults.BeaconMaxLenOsMeta).
			Optional().
			Comment("metadata of operating system"),
		field.String("hostname").
			MaxLen(defaults.BeaconMaxLenHostname).
			Optional().
			Comment("hostname of machine, on which beacon deployed"),
		field.String("username").
			MaxLen(defaults.BeaconMaxLenUsername).
			Optional().
			Comment("username of beacon's process"),
		field.String("domain").
			MaxLen(defaults.BeaconMaxLenDomain).
			Optional().
			Comment("domain of machine, on which beacon deployed"),
		field.Bool("privileged").
			Optional().
			Comment("is beacon process is privileged"),
		field.String("process_name").
			MaxLen(defaults.BeaconMaxLenProcessName).
			Optional().
			Comment("name of beacon process"),
		field.Uint32("pid").
			Optional().
			Comment("process ID of beacon"),
		field.Enum("arch").
			GoType(c.BeaconArch(0)).
			Comment("architecture of beacon process"),
		field.Uint32("sleep").
			Comment("sleep value of beacon"),
		field.Uint8("jitter").
			Comment("jitter value of sleep"),
		field.Time("first").
			Default(time.Now).
			Comment("first checkout timestamp"),
		field.Time("last").
			Default(time.Now).
			Comment("last activity of listener"),
		field.Uint32("caps").
			Comment("capabilities of beacon"),
		field.String("note").
			MaxLen(defaults.BeaconMaxLenNote).
			Optional().
			Comment("note of beacon"),
		field.Uint32("color").
			Default(defaults.DefaultColor).
			Comment("color of entity"),
	}
}

// Edges связи для Beacon
func (Beacon) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("listener", Listener.Type).
			Ref("beacon").
			Field("listener_id").
			Unique().
			Required(),
		edge.To("group", Group.Type),
		edge.To("task", Task.Type),
	}
}

// Mixin mixin для Beacon
func (Beacon) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
