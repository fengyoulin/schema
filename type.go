package schema

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Schema for a go struct
type Schema struct {
	Name   string  `json:"name,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

// Field of a struct
type Field struct {
	Name string            `json:"name,omitempty"`
	Type string            `json:"type,omitempty"` // T := basic, schema, []T, map[string]T, map[T]T, [n]T, *T
	Tags map[string]string `json:"tags,omitempty"`
}

// Types contains the basic types and schema types
type Types struct {
	tm map[string]reflect.Type
	tn map[reflect.Type]string
	lk sync.RWMutex
	os options
}

type options struct {
	disablePointer bool
	stringKeyOnly  bool
	schemaResolver Resolver
}

// Option for New
type Option func(o *options)

// Resolver func
type Resolver func(name string) (s Schema, err error)

var (
	ir *regexp.Regexp // for identifier name validation
	mr *regexp.Regexp // for map definition
	ar *regexp.Regexp // for array definition
)

func init() {
	ir = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)
	mr = regexp.MustCompile(`^map\[([A-Za-z0-9_]+)\](.+)$`)
	ar = regexp.MustCompile(`^\[([0-9]+)\](.+)$`)
}

// DisablePointer in definition
func DisablePointer() Option {
	return func(o *options) {
		o.disablePointer = true
	}
}

// StringKeyOnly in map, be compatible with json
func StringKeyOnly() Option {
	return func(o *options) {
		o.stringKeyOnly = true
	}
}

// UseResolver function
func UseResolver(r Resolver) Option {
	return func(o *options) {
		o.schemaResolver = r
	}
}

// New types and init
func New(opts ...Option) *Types {
	tm := make(map[string]reflect.Type)
	tn := make(map[reflect.Type]string)
	ts := &Types{
		tm: tm,
		tn: tn,
	}
	tm["bool"] = reflect.TypeOf(true)
	tm["int"] = reflect.TypeOf(0)
	tm["uint"] = reflect.TypeOf(uint(0))
	tm["int8"] = reflect.TypeOf(int8(0))
	tm["int16"] = reflect.TypeOf(int16(0))
	tm["int32"] = reflect.TypeOf(int32(0))
	tm["int64"] = reflect.TypeOf(int64(0))
	tm["uint8"] = reflect.TypeOf(uint8(0))
	tm["uint16"] = reflect.TypeOf(uint16(0))
	tm["uint32"] = reflect.TypeOf(uint32(0))
	tm["uint64"] = reflect.TypeOf(uint64(0))
	tm["uintptr"] = reflect.TypeOf(uintptr(0))
	tm["float32"] = reflect.TypeOf(float32(0))
	tm["float64"] = reflect.TypeOf(float64(0))
	tm["complex64"] = reflect.TypeOf(complex64(0))
	tm["complex128"] = reflect.TypeOf(complex128(0))
	tm["string"] = reflect.TypeOf("")
	for n, t := range tm {
		tn[t] = n
	}
	for _, o := range opts {
		o(&ts.os)
	}
	return ts
}

// TypeByName maps a name to type
func (ts *Types) TypeByName(name string) (t reflect.Type, ok bool) {
	ts.lk.RLock()
	t, ok = ts.tm[name]
	ts.lk.RUnlock()
	return
}

// NameByType maps a type to name
func (ts *Types) NameByType(t reflect.Type) (name string, ok bool) {
	ts.lk.RLock()
	name, ok = ts.tn[t]
	ts.lk.RUnlock()
	return
}

// CreateSchema a schema from definition
func (ts *Types) CreateSchema(s Schema) (t reflect.Type, err error) {
	ts.lk.RLock()
	t, ok := ts.tm[s.Name]
	ts.lk.RUnlock()
	if ok {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("recovered from: %v", e)
		}
	}()
	ts.lk.Lock()
	defer ts.lk.Unlock()
	return ts.createSchema(s)
}

// CreateType a type from name
func (ts *Types) CreateType(typ string) (t reflect.Type, err error) {
	ts.lk.RLock()
	t, ok := ts.tm[typ]
	ts.lk.RUnlock()
	if ok {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("recovered from: %v", e)
		}
	}()
	ts.lk.Lock()
	defer ts.lk.Unlock()
	return ts.createType(typ)
}

func (ts *Types) createSchema(s Schema) (t reflect.Type, err error) {
	if t, ok := ts.tm[s.Name]; ok {
		return t, nil
	}
	if !ir.MatchString(s.Name) {
		return nil, fmt.Errorf("invalid schema name: %s", s.Name)
	}
	fs := make([]reflect.StructField, len(s.Fields))
	for i, f := range s.Fields {
		if !ir.MatchString(f.Name) {
			return nil, fmt.Errorf("invalid field name: %s", f.Name)
		}
		t, err := ts.createType(f.Type)
		if err != nil {
			return nil, err
		}
		tags := make([]string, 0, len(f.Tags))
		for k, v := range f.Tags {
			tags = append(tags, k+`:"`+v+`"`)
		}
		sort.Strings(tags)
		fs[i] = reflect.StructField{
			Name: f.Name,
			Type: t,
			Tag:  reflect.StructTag(strings.Join(tags, " ")),
		}
	}
	t = reflect.StructOf(fs)
	ts.tm[s.Name] = t
	ts.tn[t] = s.Name
	return
}

func (ts *Types) createType(typ string) (t reflect.Type, err error) {
	t, ok := ts.tm[typ]
	if ok { // found
		return t, nil
	}
	if strings.HasPrefix(typ, "[]") { // slice
		e, err := ts.createType(typ[2:])
		if err != nil {
			return nil, err
		}
		t = reflect.SliceOf(e)
	} else if strings.HasPrefix(typ, "map[string]") { // map[string]T
		e, err := ts.createType(typ[11:])
		if err != nil {
			return nil, err
		}
		t = reflect.MapOf(reflect.TypeOf(""), e)
	} else if !ts.os.disablePointer && strings.HasPrefix(typ, "*") { // ptr
		e, err := ts.createType(typ[1:])
		if err != nil {
			return nil, err
		}
		t = reflect.PtrTo(e)
	} else if !ts.os.stringKeyOnly && mr.MatchString(typ) { // map
		ms := mr.FindStringSubmatch(typ)
		k, err := ts.createType(ms[1])
		if err != nil {
			return nil, err
		}
		v, err := ts.createType(ms[2])
		if err != nil {
			return nil, err
		}
		t = reflect.MapOf(k, v)
	} else if ar.MatchString(typ) { // array
		ms := ar.FindStringSubmatch(typ)
		n, err := strconv.ParseInt(ms[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid array type: %s, error: %v", typ, err)
		}
		e, err := ts.createType(ms[2])
		if err != nil {
			return nil, err
		}
		t = reflect.ArrayOf(int(n), e)
	} else if ts.os.schemaResolver != nil && ir.MatchString(typ) { // schema to load
		s, err := ts.os.schemaResolver(typ)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve schema: %s, error: %v", typ, err)
		}
		t, err = ts.createSchema(s)
		if err != nil {
			return nil, err
		}
	} else { // unknown
		err = fmt.Errorf("unknown type: %s", typ)
	}
	ts.tm[typ] = t
	ts.tn[t] = typ
	return
}
