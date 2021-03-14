package scanner

import (
	"fmt"
	"go/types"
)

// TODO(muvaf): Using the result of union operation as ignore func would make sense.
// Consider providing functions to make this easy. For example, `ignore all fields
// in output that already exists in input`

func WithInputFieldIgnoreFns(f ...IgnoreFieldFn) Option {
	return func(rc *RemoteCalls) {
		rc.ignore.input = f
	}
}

func WithOutputFieldIgnoreFns(f ...IgnoreFieldFn) Option {
	return func(rc *RemoteCalls) {
		rc.ignore.output = f
	}
}

func WithReadInputs(s ...string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.ReadInputs = append(r.ReadInputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithReadOutputs(s ...string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.ReadOutputs = append(r.ReadOutputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithCreateInputs(s ...string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.CreationInputs = append(r.CreationInputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithCreateOutputs(s ...string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.CreationOutputs = append(r.CreationOutputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithUpdateInputs(s ...string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.UpdateInputs = append(r.UpdateInputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithDeletionInputs(s ...string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.DeletionInputs = append(r.DeletionInputs, t.Type().(*types.Named))
			}
		}
	}
}

type Option func(*RemoteCalls)

func NewRemoteCalls(s *types.Scope, opts ...Option) *RemoteCalls {
	r := &RemoteCalls{scope: s}
	for _, f := range opts {
		f(r)
	}
	return r
}

type IgnoreFieldFn func(*types.Var) bool

type IgnoreFieldChain []IgnoreFieldFn

func (i IgnoreFieldChain) ShouldIgnore(v *types.Var) bool {
	for _, f := range i {
		if f(v) {
			return true
		}
	}
	return false
}

type ignore struct {
	input  IgnoreFieldChain
	output IgnoreFieldChain
}

type RemoteCalls struct {
	scope *types.Scope
	ignore

	CreationInputs []*types.Named
	ReadInputs     []*types.Named
	UpdateInputs   []*types.Named
	DeletionInputs []*types.Named

	CreationOutputs []*types.Named
	ReadOutputs     []*types.Named
}

func (r *RemoteCalls) AggregatedInput(tn *types.TypeName) *Named {
	varMap := map[string]*types.Var{}
	aggregatedTypes := map[string]struct{}{}
	for _, c := range r.CreationInputs {
		aggregatedTypes[fmt.Sprintf("%s.%s", c.Obj().Pkg().Path(), c.Obj().Name())] = struct{}{}
		cre := c.Underlying().(*types.Struct)
		for i := 0; i < cre.NumFields(); i++ {
			if r.ignore.input.ShouldIgnore(cre.Field(i)) {
				continue
			}
			varMap[cre.Field(i).Name()] = cre.Field(i)
		}
	}
	for _, c := range r.ReadInputs {
		aggregatedTypes[fmt.Sprintf("%s.%s", c.Obj().Pkg().Path(), c.Obj().Name())] = struct{}{}
		re := c.Underlying().(*types.Struct)
		for i := 0; i < re.NumFields(); i++ {
			if r.ignore.input.ShouldIgnore(re.Field(i)) {
				continue
			}
			varMap[re.Field(i).Name()] = re.Field(i)
		}
	}
	for _, c := range r.UpdateInputs {
		aggregatedTypes[fmt.Sprintf("%s.%s", c.Obj().Pkg().Path(), c.Obj().Name())] = struct{}{}
		u := c.Underlying().(*types.Struct)
		for i := 0; i < u.NumFields(); i++ {
			if r.ignore.input.ShouldIgnore(u.Field(i)) {
				continue
			}
			varMap[u.Field(i).Name()] = u.Field(i)
		}
	}
	for _, c := range r.DeletionInputs {
		aggregatedTypes[fmt.Sprintf("%s.%s", c.Obj().Pkg().Path(), c.Obj().Name())] = struct{}{}
		d := c.Underlying().(*types.Struct)
		for i := 0; i < d.NumFields(); i++ {
			if r.ignore.input.ShouldIgnore(d.Field(i)) {
				continue
			}
			varMap[d.Field(i).Name()] = d.Field(i)
		}
	}
	fields := make([]*types.Var, len(varMap))
	i := 0
	for _, v := range varMap {
		fields[i] = v
		i++
	}
	n := types.NewNamed(tn, types.NewStruct(fields, nil), nil)
	var ats []string
	for at := range aggregatedTypes {
		ats = append(ats, at)
	}
	return NewNamed(n, AggregatedTypesTags(ats))
}

func (r *RemoteCalls) AggregatedOutput(tn *types.TypeName) *Named {
	varMap := map[string]*types.Var{}
	aggregatedTypes := map[string]struct{}{}
	for _, c := range r.ReadOutputs {
		aggregatedTypes[fmt.Sprintf("%s.%s", c.Obj().Pkg().Path(), c.Obj().Name())] = struct{}{}
		ro := c.Underlying().(*types.Struct)
		for i := 0; i < ro.NumFields(); i++ {
			if r.ignore.output.ShouldIgnore(ro.Field(i)) {
				continue
			}
			varMap[ro.Field(i).Name()] = ro.Field(i)
		}
	}
	for _, c := range r.CreationOutputs {
		aggregatedTypes[fmt.Sprintf("%s.%s", c.Obj().Pkg().Path(), c.Obj().Name())] = struct{}{}
		co := c.Underlying().(*types.Struct)
		for i := 0; i < co.NumFields(); i++ {
			if r.ignore.output.ShouldIgnore(co.Field(i)) {
				continue
			}
			varMap[co.Field(i).Name()] = co.Field(i)
		}
	}
	fields := make([]*types.Var, len(varMap))
	i := 0
	for _, v := range varMap {
		fields[i] = v
		i++
	}
	n := types.NewNamed(tn, types.NewStruct(fields, nil), nil)
	var ats []string
	for at := range aggregatedTypes {
		ats = append(ats, at)
	}
	return NewNamed(n, AggregatedTypesTags(ats))
}
