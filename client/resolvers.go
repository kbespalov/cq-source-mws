package client

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/thoas/go-funk"
)

// A converter turns one value read off a model into what its column holds.
// Returning a nil value leaves the column null.
//
// Conversion is kept apart from reading because the same value can be reached
// two ways: as a field of the resource, and as an element of a slice, where
// there is no path to read from. Both need the identical treatment, or a list
// of addresses would end up spelled differently from an address of its own.
type converter func(data any) (any, error)

// resolveValue reads the field at path and stores what the converter makes of
// it. Paths are dotted and read with funk.Get, which yields nil for anything
// missing on the way, so an absent field leaves the column null.
func resolveValue(path string, value converter) schema.ColumnResolver {
	return func(_ context.Context, _ schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
		data := funk.Get(resource.Item, path)
		if data == nil {
			return nil
		}
		converted, err := value(data)
		if err != nil {
			return fmt.Errorf("column %q: %w", c.Name, err)
		}
		if converted == nil {
			return nil
		}
		return resource.Set(c.Name, converted)
	}
}

// resourceIDValue is the full resource path of an AnyResourceID or a typed
// *ID, e.g. "rm/projects/my-project".
func resourceIDValue(data any) (any, error) {
	id, ok := addr(data).(resourceID)
	if !ok {
		return nil, fmt.Errorf("wanted a resource id, have %T", data)
	}
	return id.ID(), nil
}

// refValue is the absolute path a reference points at, e.g.
// "org/organizations/my-org/folders/my-folder".
//
// A reference carries the path twice. Path() is what the server actually wrote
// and may be relative: the same organization comes back as
// "org/organizations/my-org" from one field and "organizations/my-org" from
// another, purely depending on the caller's context. IDPath() normalises that
// to the absolute, service-qualified form, which is the only one that joins
// across tables. Path() remains the fallback for a relative reference that was
// never resolved, where IDPath() has nothing to build an absolute path from.
func refValue(data any) (any, error) {
	ref, ok := addr(data).(resourceRef)
	if !ok {
		return nil, fmt.Errorf("wanted a resource reference, have %T", data)
	}
	return RefPath(ref), nil
}

// RefPath returns the absolute path of a reference, falling back to the
// original one when the reference is relative and unresolved.
func RefPath(ref resourceRef) string {
	if absolute := ref.IDPath(); absolute != "" {
		return absolute
	}
	return ref.Path()
}

// quantityValue is the amount a unit type holds, in the base unit of its
// dimension: bytes for a size, hertz for a frequency. The unit the API wrote
// it in is lost, which is the point: only one scale can be summed.
func quantityValue(data any) (any, error) {
	amount, ok := addr(data).(quantity)
	if !ok {
		return nil, fmt.Errorf("wanted a quantity, have %T", data)
	}
	base := amount.BigInt()
	if base == nil {
		return nil, nil
	}
	if !base.IsInt64() {
		return nil, fmt.Errorf("quantity %s does not fit a 64-bit column", base)
	}
	return base.Int64(), nil
}

// inetValue is an address, handed over as the text it was written in.
//
// The CloudQuery scalar parses that text itself and normalises an IPv4 address
// to its four-byte form on the way. Passing the net.IP out of the model would
// skip that step: net.ParseIP keeps every address 16 bytes wide, so 10.0.0.1
// would be stored as ::ffff:10.0.0.1/128.
func inetValue(data any) (any, error) {
	value := addr(data)
	_, cidr := value.(cidrAddress)
	_, ip := value.(ipAddress)
	if !cidr && !ip {
		return nil, fmt.Errorf("wanted an address, have %T", data)
	}
	return stringerValue(data)
}

// stringerValue is the rendering of a type that keeps its state unexported,
// such as a duration written back as "PT30S".
func stringerValue(data any) (any, error) {
	s, ok := addr(data).(fmt.Stringer)
	if !ok {
		return nil, fmt.Errorf("wanted a fmt.Stringer, have %T", data)
	}
	value := s.String()
	if value == "" {
		return nil, nil
	}
	return value, nil
}

// optionalValue unwraps optional.Optional[T] and optional.OptionalNil[T],
// leaving the column null when the field was not set.
func optionalValue(inner converter) converter {
	return func(data any) (any, error) {
		v := reflect.Indirect(reflect.ValueOf(data))
		if set := v.FieldByName("Set"); set.IsValid() && !set.Bool() {
			return nil, nil
		}
		if null := v.FieldByName("Null"); null.IsValid() && null.Bool() {
			return nil, nil
		}
		value := v.FieldByName("Value")
		if !value.IsValid() {
			return nil, fmt.Errorf("wanted an optional, have %T", data)
		}
		if value.Kind() == reflect.Pointer && value.IsNil() {
			return nil, nil
		}
		if inner == nil {
			return value.Interface(), nil
		}
		return inner(value.Interface())
	}
}

// sliceValue converts a slice element by element, so that a list column holds
// the same values a column of one element would.
func sliceValue(elem converter) converter {
	return func(data any) (any, error) {
		v := reflect.Indirect(reflect.ValueOf(data))
		if v.Kind() != reflect.Slice {
			return nil, fmt.Errorf("wanted a slice, have %T", data)
		}
		if v.Len() == 0 {
			return nil, nil
		}
		values := make([]any, v.Len())
		for i := range values {
			converted, err := elem(v.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			values[i] = converted
		}
		return values, nil
	}
}

// ResolveResourceName stores just the resource name of an identifier, e.g.
// "my-vm" for "compute/projects/my-project/virtualMachines/my-vm". The SDK
// derives it from the parsed identifier, so it stays correct for paths whose
// depth varies between services.
func ResolveResourceName(path string) schema.ColumnResolver {
	return resolveValue(path, resourceNameValue)
}

// resourceNameValue is the trailing name of an identifier. Nearly every
// service writes metadata.id as a resource id, which parses itself; regions
// and zones write theirs as a reference instead, and there the name is the
// last segment of the path.
func resourceNameValue(data any) (any, error) {
	switch value := addr(data).(type) {
	case resourceID:
		return string(value.ResourceName()), nil
	case resourceRef:
		return lastSegment(RefPath(value)), nil
	}
	return nil, fmt.Errorf("wanted a resource id or reference, have %T", data)
}

// addr returns a pointer to data so that methods with pointer receivers, which
// is how the SDK declares them, are reachable.
func addr(data any) any {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		return data
	}
	p := reflect.New(v.Type())
	p.Elem().Set(v)
	return p.Interface()
}

// ResolveOrganization, ResolveFolder and ResolveProject fill the hierarchy
// columns every table carries.
func ResolveOrganization(_ context.Context, meta schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
	return resource.Set(c.Name, meta.(*Client).OrganizationID)
}

func ResolveFolder(_ context.Context, meta schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
	return resource.Set(c.Name, meta.(*Client).FolderID)
}

func ResolveProject(_ context.Context, meta schema.ClientMeta, resource *schema.Resource, c schema.Column) error {
	return resource.Set(c.Name, meta.(*Client).ProjectID)
}
