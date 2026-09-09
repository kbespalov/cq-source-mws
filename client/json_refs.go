package client

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	valuesctx "go.mws.cloud/go-sdk/pkg/context/values"
)

// jsonAbsoluteRefsValue stores a model JSON column the way json.Marshal would,
// then rewrites every nested resource id and reference to the absolute path
// the scalar columns already use. The SDK's own Encode writes Path(), which
// is whatever the server sent and may be relative.
func jsonAbsoluteRefsValue(ctx context.Context, c *Client, data any) (any, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return nil, nil
	}
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, err
	}
	rewritten := rewriteJSONRefs(ctx, c, reflect.ValueOf(data), tree)
	out, err := json.Marshal(rewritten)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

// rewriteJSONRefs walks a Go value and the JSON it produced in lockstep and
// replaces every resource id or reference string with its absolute path.
func rewriteJSONRefs(ctx context.Context, c *Client, v reflect.Value, tree any) any {
	if tree == nil {
		return nil
	}
	v = unwrapJSONValue(v)
	if !v.IsValid() {
		return tree
	}

	if v.CanInterface() {
		data := addr(v.Interface())
		if id, ok := data.(resourceID); ok {
			if _, isStr := tree.(string); isStr {
				if abs := id.ID(); abs != "" {
					return abs
				}
				return tree
			}
		}
		if ref, ok := data.(resourceRef); ok {
			if _, isStr := tree.(string); isStr {
				return absoluteRef(ctx, c, ref)
			}
		}
	}

	switch node := tree.(type) {
	case map[string]any:
		return rewriteJSONObject(ctx, c, v, node)
	case []any:
		return rewriteJSONArray(ctx, c, v, node)
	default:
		return tree
	}
}

func rewriteJSONObject(ctx context.Context, c *Client, v reflect.Value, tree map[string]any) map[string]any {
	if v.Kind() != reflect.Struct {
		return tree
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, skip, inline := jsonField(field)
		if skip {
			continue
		}
		fv := v.Field(i)
		if inline {
			rewriteJSONObject(ctx, c, unwrapJSONValue(fv), tree)
			continue
		}
		child, ok := tree[name]
		if !ok {
			continue
		}
		tree[name] = rewriteJSONRefs(ctx, c, fv, child)
	}
	return tree
}

func rewriteJSONArray(ctx context.Context, c *Client, v reflect.Value, tree []any) []any {
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return tree
	}
	n := v.Len()
	if n > len(tree) {
		n = len(tree)
	}
	for i := 0; i < n; i++ {
		tree[i] = rewriteJSONRefs(ctx, c, v.Index(i), tree[i])
	}
	return tree
}

// unwrapJSONValue peels pointers, interfaces and optionals. The SDK marshals
// an optional as the value it holds, so the JSON tree is already unwrapped
// and the walk has to match.
func unwrapJSONValue(v reflect.Value) reflect.Value {
	for v.IsValid() {
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
			continue
		case reflect.Struct:
			if _, ok := unwrapOptional(v.Type()); !ok {
				return v
			}
			if set := v.FieldByName("Set"); set.IsValid() && !set.Bool() {
				return reflect.Value{}
			}
			if null := v.FieldByName("Null"); null.IsValid() && null.Bool() {
				return reflect.Value{}
			}
			v = v.FieldByName("Value")
			continue
		}
		return v
	}
	return v
}

func jsonField(field reflect.StructField) (name string, skip bool, inline bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", true, false
	}
	name, _, _ = strings.Cut(tag, ",")
	if field.Anonymous && name == "" {
		return "", false, true
	}
	if name == "" {
		name = field.Name
	}
	return name, false, false
}

// absoluteRef is the path a reference should be stored as. IDPath() is the
// service-qualified form that joins to metadata.id. When the server wrote
// only a name, Parse can fill the missing segments from the multiplex
// project; if that still leaves nothing, Path() is what arrived on the wire.
func absoluteRef(ctx context.Context, c *Client, ref resourceRef) string {
	if abs := ref.IDPath(); abs != "" {
		return abs
	}
	if c != nil && c.ProjectName != "" {
		if p, ok := addr(ref).(interface{ Parse(context.Context) error }); ok {
			_ = p.Parse(valuesctx.With(ctx, "project", c.ProjectName))
			if abs := ref.IDPath(); abs != "" {
				return abs
			}
		}
	}
	return ref.Path()
}
