// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/c2micro/c2msrv/internal/ent/beacon"
	"github.com/c2micro/c2msrv/internal/ent/group"
	"github.com/c2micro/c2msrv/internal/ent/listener"
	"github.com/c2micro/c2msrv/internal/ent/task"
	"github.com/c2micro/c2msrv/internal/types"
	"github.com/c2micro/c2mshr/defaults"
)

// BeaconCreate is the builder for creating a Beacon entity.
type BeaconCreate struct {
	config
	mutation *BeaconMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (bc *BeaconCreate) SetCreatedAt(t time.Time) *BeaconCreate {
	bc.mutation.SetCreatedAt(t)
	return bc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableCreatedAt(t *time.Time) *BeaconCreate {
	if t != nil {
		bc.SetCreatedAt(*t)
	}
	return bc
}

// SetUpdatedAt sets the "updated_at" field.
func (bc *BeaconCreate) SetUpdatedAt(t time.Time) *BeaconCreate {
	bc.mutation.SetUpdatedAt(t)
	return bc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableUpdatedAt(t *time.Time) *BeaconCreate {
	if t != nil {
		bc.SetUpdatedAt(*t)
	}
	return bc
}

// SetDeletedAt sets the "deleted_at" field.
func (bc *BeaconCreate) SetDeletedAt(t time.Time) *BeaconCreate {
	bc.mutation.SetDeletedAt(t)
	return bc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableDeletedAt(t *time.Time) *BeaconCreate {
	if t != nil {
		bc.SetDeletedAt(*t)
	}
	return bc
}

// SetBid sets the "bid" field.
func (bc *BeaconCreate) SetBid(u uint32) *BeaconCreate {
	bc.mutation.SetBid(u)
	return bc
}

// SetListenerID sets the "listener_id" field.
func (bc *BeaconCreate) SetListenerID(i int) *BeaconCreate {
	bc.mutation.SetListenerID(i)
	return bc
}

// SetExtIP sets the "ext_ip" field.
func (bc *BeaconCreate) SetExtIP(t types.Inet) *BeaconCreate {
	bc.mutation.SetExtIP(t)
	return bc
}

// SetNillableExtIP sets the "ext_ip" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableExtIP(t *types.Inet) *BeaconCreate {
	if t != nil {
		bc.SetExtIP(*t)
	}
	return bc
}

// SetIntIP sets the "int_ip" field.
func (bc *BeaconCreate) SetIntIP(t types.Inet) *BeaconCreate {
	bc.mutation.SetIntIP(t)
	return bc
}

// SetNillableIntIP sets the "int_ip" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableIntIP(t *types.Inet) *BeaconCreate {
	if t != nil {
		bc.SetIntIP(*t)
	}
	return bc
}

// SetOs sets the "os" field.
func (bc *BeaconCreate) SetOs(do defaults.BeaconOS) *BeaconCreate {
	bc.mutation.SetOs(do)
	return bc
}

// SetOsMeta sets the "os_meta" field.
func (bc *BeaconCreate) SetOsMeta(s string) *BeaconCreate {
	bc.mutation.SetOsMeta(s)
	return bc
}

// SetNillableOsMeta sets the "os_meta" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableOsMeta(s *string) *BeaconCreate {
	if s != nil {
		bc.SetOsMeta(*s)
	}
	return bc
}

// SetHostname sets the "hostname" field.
func (bc *BeaconCreate) SetHostname(s string) *BeaconCreate {
	bc.mutation.SetHostname(s)
	return bc
}

// SetNillableHostname sets the "hostname" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableHostname(s *string) *BeaconCreate {
	if s != nil {
		bc.SetHostname(*s)
	}
	return bc
}

// SetUsername sets the "username" field.
func (bc *BeaconCreate) SetUsername(s string) *BeaconCreate {
	bc.mutation.SetUsername(s)
	return bc
}

// SetNillableUsername sets the "username" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableUsername(s *string) *BeaconCreate {
	if s != nil {
		bc.SetUsername(*s)
	}
	return bc
}

// SetDomain sets the "domain" field.
func (bc *BeaconCreate) SetDomain(s string) *BeaconCreate {
	bc.mutation.SetDomain(s)
	return bc
}

// SetNillableDomain sets the "domain" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableDomain(s *string) *BeaconCreate {
	if s != nil {
		bc.SetDomain(*s)
	}
	return bc
}

// SetPrivileged sets the "privileged" field.
func (bc *BeaconCreate) SetPrivileged(b bool) *BeaconCreate {
	bc.mutation.SetPrivileged(b)
	return bc
}

// SetNillablePrivileged sets the "privileged" field if the given value is not nil.
func (bc *BeaconCreate) SetNillablePrivileged(b *bool) *BeaconCreate {
	if b != nil {
		bc.SetPrivileged(*b)
	}
	return bc
}

// SetProcessName sets the "process_name" field.
func (bc *BeaconCreate) SetProcessName(s string) *BeaconCreate {
	bc.mutation.SetProcessName(s)
	return bc
}

// SetNillableProcessName sets the "process_name" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableProcessName(s *string) *BeaconCreate {
	if s != nil {
		bc.SetProcessName(*s)
	}
	return bc
}

// SetPid sets the "pid" field.
func (bc *BeaconCreate) SetPid(u uint32) *BeaconCreate {
	bc.mutation.SetPid(u)
	return bc
}

// SetNillablePid sets the "pid" field if the given value is not nil.
func (bc *BeaconCreate) SetNillablePid(u *uint32) *BeaconCreate {
	if u != nil {
		bc.SetPid(*u)
	}
	return bc
}

// SetArch sets the "arch" field.
func (bc *BeaconCreate) SetArch(da defaults.BeaconArch) *BeaconCreate {
	bc.mutation.SetArch(da)
	return bc
}

// SetSleep sets the "sleep" field.
func (bc *BeaconCreate) SetSleep(u uint32) *BeaconCreate {
	bc.mutation.SetSleep(u)
	return bc
}

// SetJitter sets the "jitter" field.
func (bc *BeaconCreate) SetJitter(u uint8) *BeaconCreate {
	bc.mutation.SetJitter(u)
	return bc
}

// SetFirst sets the "first" field.
func (bc *BeaconCreate) SetFirst(t time.Time) *BeaconCreate {
	bc.mutation.SetFirst(t)
	return bc
}

// SetNillableFirst sets the "first" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableFirst(t *time.Time) *BeaconCreate {
	if t != nil {
		bc.SetFirst(*t)
	}
	return bc
}

// SetLast sets the "last" field.
func (bc *BeaconCreate) SetLast(t time.Time) *BeaconCreate {
	bc.mutation.SetLast(t)
	return bc
}

// SetNillableLast sets the "last" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableLast(t *time.Time) *BeaconCreate {
	if t != nil {
		bc.SetLast(*t)
	}
	return bc
}

// SetCaps sets the "caps" field.
func (bc *BeaconCreate) SetCaps(u uint32) *BeaconCreate {
	bc.mutation.SetCaps(u)
	return bc
}

// SetNote sets the "note" field.
func (bc *BeaconCreate) SetNote(s string) *BeaconCreate {
	bc.mutation.SetNote(s)
	return bc
}

// SetNillableNote sets the "note" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableNote(s *string) *BeaconCreate {
	if s != nil {
		bc.SetNote(*s)
	}
	return bc
}

// SetColor sets the "color" field.
func (bc *BeaconCreate) SetColor(u uint32) *BeaconCreate {
	bc.mutation.SetColor(u)
	return bc
}

// SetNillableColor sets the "color" field if the given value is not nil.
func (bc *BeaconCreate) SetNillableColor(u *uint32) *BeaconCreate {
	if u != nil {
		bc.SetColor(*u)
	}
	return bc
}

// SetListener sets the "listener" edge to the Listener entity.
func (bc *BeaconCreate) SetListener(l *Listener) *BeaconCreate {
	return bc.SetListenerID(l.ID)
}

// AddGroupIDs adds the "group" edge to the Group entity by IDs.
func (bc *BeaconCreate) AddGroupIDs(ids ...int) *BeaconCreate {
	bc.mutation.AddGroupIDs(ids...)
	return bc
}

// AddGroup adds the "group" edges to the Group entity.
func (bc *BeaconCreate) AddGroup(g ...*Group) *BeaconCreate {
	ids := make([]int, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return bc.AddGroupIDs(ids...)
}

// AddTaskIDs adds the "task" edge to the Task entity by IDs.
func (bc *BeaconCreate) AddTaskIDs(ids ...int) *BeaconCreate {
	bc.mutation.AddTaskIDs(ids...)
	return bc
}

// AddTask adds the "task" edges to the Task entity.
func (bc *BeaconCreate) AddTask(t ...*Task) *BeaconCreate {
	ids := make([]int, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return bc.AddTaskIDs(ids...)
}

// Mutation returns the BeaconMutation object of the builder.
func (bc *BeaconCreate) Mutation() *BeaconMutation {
	return bc.mutation
}

// Save creates the Beacon in the database.
func (bc *BeaconCreate) Save(ctx context.Context) (*Beacon, error) {
	if err := bc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, bc.sqlSave, bc.mutation, bc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (bc *BeaconCreate) SaveX(ctx context.Context) *Beacon {
	v, err := bc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (bc *BeaconCreate) Exec(ctx context.Context) error {
	_, err := bc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bc *BeaconCreate) ExecX(ctx context.Context) {
	if err := bc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (bc *BeaconCreate) defaults() error {
	if _, ok := bc.mutation.CreatedAt(); !ok {
		if beacon.DefaultCreatedAt == nil {
			return fmt.Errorf("ent: uninitialized beacon.DefaultCreatedAt (forgotten import ent/runtime?)")
		}
		v := beacon.DefaultCreatedAt()
		bc.mutation.SetCreatedAt(v)
	}
	if _, ok := bc.mutation.UpdatedAt(); !ok {
		if beacon.DefaultUpdatedAt == nil {
			return fmt.Errorf("ent: uninitialized beacon.DefaultUpdatedAt (forgotten import ent/runtime?)")
		}
		v := beacon.DefaultUpdatedAt()
		bc.mutation.SetUpdatedAt(v)
	}
	if _, ok := bc.mutation.First(); !ok {
		if beacon.DefaultFirst == nil {
			return fmt.Errorf("ent: uninitialized beacon.DefaultFirst (forgotten import ent/runtime?)")
		}
		v := beacon.DefaultFirst()
		bc.mutation.SetFirst(v)
	}
	if _, ok := bc.mutation.Last(); !ok {
		if beacon.DefaultLast == nil {
			return fmt.Errorf("ent: uninitialized beacon.DefaultLast (forgotten import ent/runtime?)")
		}
		v := beacon.DefaultLast()
		bc.mutation.SetLast(v)
	}
	if _, ok := bc.mutation.Color(); !ok {
		v := beacon.DefaultColor
		bc.mutation.SetColor(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (bc *BeaconCreate) check() error {
	if _, ok := bc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Beacon.created_at"`)}
	}
	if _, ok := bc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Beacon.updated_at"`)}
	}
	if _, ok := bc.mutation.Bid(); !ok {
		return &ValidationError{Name: "bid", err: errors.New(`ent: missing required field "Beacon.bid"`)}
	}
	if _, ok := bc.mutation.ListenerID(); !ok {
		return &ValidationError{Name: "listener_id", err: errors.New(`ent: missing required field "Beacon.listener_id"`)}
	}
	if v, ok := bc.mutation.ExtIP(); ok {
		if err := beacon.ExtIPValidator(v.String()); err != nil {
			return &ValidationError{Name: "ext_ip", err: fmt.Errorf(`ent: validator failed for field "Beacon.ext_ip": %w`, err)}
		}
	}
	if v, ok := bc.mutation.IntIP(); ok {
		if err := beacon.IntIPValidator(v.String()); err != nil {
			return &ValidationError{Name: "int_ip", err: fmt.Errorf(`ent: validator failed for field "Beacon.int_ip": %w`, err)}
		}
	}
	if _, ok := bc.mutation.Os(); !ok {
		return &ValidationError{Name: "os", err: errors.New(`ent: missing required field "Beacon.os"`)}
	}
	if v, ok := bc.mutation.Os(); ok {
		if err := beacon.OsValidator(v); err != nil {
			return &ValidationError{Name: "os", err: fmt.Errorf(`ent: validator failed for field "Beacon.os": %w`, err)}
		}
	}
	if v, ok := bc.mutation.OsMeta(); ok {
		if err := beacon.OsMetaValidator(v); err != nil {
			return &ValidationError{Name: "os_meta", err: fmt.Errorf(`ent: validator failed for field "Beacon.os_meta": %w`, err)}
		}
	}
	if v, ok := bc.mutation.Hostname(); ok {
		if err := beacon.HostnameValidator(v); err != nil {
			return &ValidationError{Name: "hostname", err: fmt.Errorf(`ent: validator failed for field "Beacon.hostname": %w`, err)}
		}
	}
	if v, ok := bc.mutation.Username(); ok {
		if err := beacon.UsernameValidator(v); err != nil {
			return &ValidationError{Name: "username", err: fmt.Errorf(`ent: validator failed for field "Beacon.username": %w`, err)}
		}
	}
	if v, ok := bc.mutation.Domain(); ok {
		if err := beacon.DomainValidator(v); err != nil {
			return &ValidationError{Name: "domain", err: fmt.Errorf(`ent: validator failed for field "Beacon.domain": %w`, err)}
		}
	}
	if v, ok := bc.mutation.ProcessName(); ok {
		if err := beacon.ProcessNameValidator(v); err != nil {
			return &ValidationError{Name: "process_name", err: fmt.Errorf(`ent: validator failed for field "Beacon.process_name": %w`, err)}
		}
	}
	if _, ok := bc.mutation.Arch(); !ok {
		return &ValidationError{Name: "arch", err: errors.New(`ent: missing required field "Beacon.arch"`)}
	}
	if v, ok := bc.mutation.Arch(); ok {
		if err := beacon.ArchValidator(v); err != nil {
			return &ValidationError{Name: "arch", err: fmt.Errorf(`ent: validator failed for field "Beacon.arch": %w`, err)}
		}
	}
	if _, ok := bc.mutation.Sleep(); !ok {
		return &ValidationError{Name: "sleep", err: errors.New(`ent: missing required field "Beacon.sleep"`)}
	}
	if _, ok := bc.mutation.Jitter(); !ok {
		return &ValidationError{Name: "jitter", err: errors.New(`ent: missing required field "Beacon.jitter"`)}
	}
	if _, ok := bc.mutation.First(); !ok {
		return &ValidationError{Name: "first", err: errors.New(`ent: missing required field "Beacon.first"`)}
	}
	if _, ok := bc.mutation.Last(); !ok {
		return &ValidationError{Name: "last", err: errors.New(`ent: missing required field "Beacon.last"`)}
	}
	if _, ok := bc.mutation.Caps(); !ok {
		return &ValidationError{Name: "caps", err: errors.New(`ent: missing required field "Beacon.caps"`)}
	}
	if v, ok := bc.mutation.Note(); ok {
		if err := beacon.NoteValidator(v); err != nil {
			return &ValidationError{Name: "note", err: fmt.Errorf(`ent: validator failed for field "Beacon.note": %w`, err)}
		}
	}
	if _, ok := bc.mutation.Color(); !ok {
		return &ValidationError{Name: "color", err: errors.New(`ent: missing required field "Beacon.color"`)}
	}
	if len(bc.mutation.ListenerIDs()) == 0 {
		return &ValidationError{Name: "listener", err: errors.New(`ent: missing required edge "Beacon.listener"`)}
	}
	return nil
}

func (bc *BeaconCreate) sqlSave(ctx context.Context) (*Beacon, error) {
	if err := bc.check(); err != nil {
		return nil, err
	}
	_node, _spec := bc.createSpec()
	if err := sqlgraph.CreateNode(ctx, bc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	bc.mutation.id = &_node.ID
	bc.mutation.done = true
	return _node, nil
}

func (bc *BeaconCreate) createSpec() (*Beacon, *sqlgraph.CreateSpec) {
	var (
		_node = &Beacon{config: bc.config}
		_spec = sqlgraph.NewCreateSpec(beacon.Table, sqlgraph.NewFieldSpec(beacon.FieldID, field.TypeInt))
	)
	if value, ok := bc.mutation.CreatedAt(); ok {
		_spec.SetField(beacon.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := bc.mutation.UpdatedAt(); ok {
		_spec.SetField(beacon.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := bc.mutation.DeletedAt(); ok {
		_spec.SetField(beacon.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := bc.mutation.Bid(); ok {
		_spec.SetField(beacon.FieldBid, field.TypeUint32, value)
		_node.Bid = value
	}
	if value, ok := bc.mutation.ExtIP(); ok {
		_spec.SetField(beacon.FieldExtIP, field.TypeString, value)
		_node.ExtIP = value
	}
	if value, ok := bc.mutation.IntIP(); ok {
		_spec.SetField(beacon.FieldIntIP, field.TypeString, value)
		_node.IntIP = value
	}
	if value, ok := bc.mutation.Os(); ok {
		_spec.SetField(beacon.FieldOs, field.TypeEnum, value)
		_node.Os = value
	}
	if value, ok := bc.mutation.OsMeta(); ok {
		_spec.SetField(beacon.FieldOsMeta, field.TypeString, value)
		_node.OsMeta = value
	}
	if value, ok := bc.mutation.Hostname(); ok {
		_spec.SetField(beacon.FieldHostname, field.TypeString, value)
		_node.Hostname = value
	}
	if value, ok := bc.mutation.Username(); ok {
		_spec.SetField(beacon.FieldUsername, field.TypeString, value)
		_node.Username = value
	}
	if value, ok := bc.mutation.Domain(); ok {
		_spec.SetField(beacon.FieldDomain, field.TypeString, value)
		_node.Domain = value
	}
	if value, ok := bc.mutation.Privileged(); ok {
		_spec.SetField(beacon.FieldPrivileged, field.TypeBool, value)
		_node.Privileged = value
	}
	if value, ok := bc.mutation.ProcessName(); ok {
		_spec.SetField(beacon.FieldProcessName, field.TypeString, value)
		_node.ProcessName = value
	}
	if value, ok := bc.mutation.Pid(); ok {
		_spec.SetField(beacon.FieldPid, field.TypeUint32, value)
		_node.Pid = value
	}
	if value, ok := bc.mutation.Arch(); ok {
		_spec.SetField(beacon.FieldArch, field.TypeEnum, value)
		_node.Arch = value
	}
	if value, ok := bc.mutation.Sleep(); ok {
		_spec.SetField(beacon.FieldSleep, field.TypeUint32, value)
		_node.Sleep = value
	}
	if value, ok := bc.mutation.Jitter(); ok {
		_spec.SetField(beacon.FieldJitter, field.TypeUint8, value)
		_node.Jitter = value
	}
	if value, ok := bc.mutation.First(); ok {
		_spec.SetField(beacon.FieldFirst, field.TypeTime, value)
		_node.First = value
	}
	if value, ok := bc.mutation.Last(); ok {
		_spec.SetField(beacon.FieldLast, field.TypeTime, value)
		_node.Last = value
	}
	if value, ok := bc.mutation.Caps(); ok {
		_spec.SetField(beacon.FieldCaps, field.TypeUint32, value)
		_node.Caps = value
	}
	if value, ok := bc.mutation.Note(); ok {
		_spec.SetField(beacon.FieldNote, field.TypeString, value)
		_node.Note = value
	}
	if value, ok := bc.mutation.Color(); ok {
		_spec.SetField(beacon.FieldColor, field.TypeUint32, value)
		_node.Color = value
	}
	if nodes := bc.mutation.ListenerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   beacon.ListenerTable,
			Columns: []string{beacon.ListenerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(listener.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ListenerID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := bc.mutation.GroupIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   beacon.GroupTable,
			Columns: []string{beacon.GroupColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := bc.mutation.TaskIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   beacon.TaskTable,
			Columns: []string{beacon.TaskColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(task.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// BeaconCreateBulk is the builder for creating many Beacon entities in bulk.
type BeaconCreateBulk struct {
	config
	err      error
	builders []*BeaconCreate
}

// Save creates the Beacon entities in the database.
func (bcb *BeaconCreateBulk) Save(ctx context.Context) ([]*Beacon, error) {
	if bcb.err != nil {
		return nil, bcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(bcb.builders))
	nodes := make([]*Beacon, len(bcb.builders))
	mutators := make([]Mutator, len(bcb.builders))
	for i := range bcb.builders {
		func(i int, root context.Context) {
			builder := bcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*BeaconMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, bcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, bcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, bcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (bcb *BeaconCreateBulk) SaveX(ctx context.Context) []*Beacon {
	v, err := bcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (bcb *BeaconCreateBulk) Exec(ctx context.Context) error {
	_, err := bcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bcb *BeaconCreateBulk) ExecX(ctx context.Context) {
	if err := bcb.Exec(ctx); err != nil {
		panic(err)
	}
}