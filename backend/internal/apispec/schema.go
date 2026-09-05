// Package apispec derives the API contract from the Go types the handlers
// actually return.
//
// The alternative — hand-writing an OpenAPI document, or hand-writing Dart and
// TypeScript models — means three descriptions of the same thing that drift the
// moment someone adds a field. Reflecting over the real structs makes the Go
// code the single source of truth, and `make apigen` plus a staleness test keep
// the generated clients honest.
package apispec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Schema is the subset of JSON Schema the generators need.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Nullable    bool               `json:"nullable,omitempty"`
	Description string             `json:"description,omitempty"`
	// AdditionalProperties models map[string]T.
	AdditionalProperties *Schema `json:"additionalProperties,omitempty"`
}

// Registry collects named schemas discovered by reflection.
type Registry struct {
	Schemas map[string]*Schema
	order   []string
	// names maps a Go type to the schema name chosen for it, so a second type
	// with a clashing short name gets the package-qualified form instead.
	names map[reflect.Type]string
	taken map[string]reflect.Type
}

func NewRegistry() *Registry {
	return &Registry{
		Schemas: map[string]*Schema{},
		names:   map[reflect.Type]string{},
		taken:   map[string]reflect.Type{},
	}
}

// nameFor picks the friendliest unique name for a type.
//
// The package prefix is dropped when the type name already carries it
// (project.Project reads better as Project than ProjectProject) and when the
// type is one of this package's own response envelopes. Anything that would
// collide falls back to the qualified name.
func (r *Registry) nameFor(t reflect.Type) string {
	if name, ok := r.names[t]; ok {
		return name
	}
	pkg := packageName(t)
	base := title(t.Name())

	candidates := []string{}
	if pkg == "apispec" || strings.EqualFold(base, pkg) {
		candidates = append(candidates, base)
	}
	candidates = append(candidates, title(pkg)+base)

	chosen := candidates[len(candidates)-1]
	for _, c := range candidates {
		if owner, used := r.taken[c]; !used || owner == t {
			chosen = c
			break
		}
	}
	r.names[t] = chosen
	r.taken[chosen] = t
	return chosen
}

func packageName(t reflect.Type) string {
	pkg := t.PkgPath()
	if i := strings.LastIndex(pkg, "/"); i >= 0 {
		pkg = pkg[i+1:]
	}
	return pkg
}

// Names returns registered schema names in a stable order.
func (r *Registry) Names() []string {
	out := append([]string(nil), r.order...)
	sort.Strings(out)
	return out
}

var (
	timeType = reflect.TypeOf(time.Time{})
	rawType  = reflect.TypeOf(json.RawMessage{})
)

// Register walks t and returns a schema referencing it, adding every named
// struct it meets to the registry.
func (r *Registry) Register(t reflect.Type) *Schema {
	return r.schemaFor(t, map[reflect.Type]bool{})
}

func (r *Registry) schemaFor(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch {
	case t == timeType:
		return &Schema{Type: "string", Format: "date-time"}
	case t == rawType:
		// Free-form JSON: the dynamic form schema and project context.
		return &Schema{Type: "object"}
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return &Schema{Type: "string", Format: "byte"}
		}
		return &Schema{Type: "array", Items: r.schemaFor(t.Elem(), seen)}
	case reflect.Map:
		return &Schema{Type: "object", AdditionalProperties: r.schemaFor(t.Elem(), seen)}
	case reflect.Interface:
		return &Schema{} // any
	case reflect.Struct:
		return r.structSchema(t, seen)
	}
	return &Schema{}
}

func (r *Registry) structSchema(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	if t.Name() == "" {
		return r.inlineStruct(t, seen) // anonymous struct
	}
	name := r.nameFor(t)
	ref := &Schema{Ref: "#/components/schemas/" + name}
	if _, done := r.Schemas[name]; done || seen[t] {
		return ref
	}
	seen[t] = true
	r.Schemas[name] = r.inlineStruct(t, seen)
	r.order = append(r.order, name)
	return ref
}

func (r *Registry) inlineStruct(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	s := &Schema{Type: "object", Properties: map[string]*Schema{}}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue // unexported
		}
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		jsonName, opts, _ := strings.Cut(tag, ",")
		if jsonName == "" {
			jsonName = f.Name
		}
		// Embedded structs without a json name flatten into the parent.
		if f.Anonymous && jsonName == f.Name {
			embedded := r.inlineStruct(derefType(f.Type), seen)
			for k, v := range embedded.Properties {
				s.Properties[k] = v
			}
			s.Required = append(s.Required, embedded.Required...)
			continue
		}
		field := r.schemaFor(f.Type, seen)
		if f.Type.Kind() == reflect.Ptr {
			field.Nullable = true
		}
		s.Properties[jsonName] = field
		if !strings.Contains(opts, "omitempty") && f.Type.Kind() != reflect.Ptr {
			s.Required = append(s.Required, jsonName)
		}
	}
	sort.Strings(s.Required)
	return s
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (s *Schema) String() string { return fmt.Sprintf("%+v", *s) }

// TypeName is the registry name chosen for a value's type.
func (r *Registry) TypeName(v any) string {
	return r.nameFor(derefType(reflect.TypeOf(v)))
}

func errorResponseType() reflect.Type { return reflect.TypeOf(ErrorResponse{}) }
