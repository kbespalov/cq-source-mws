package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"reflect"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	cqtypes "github.com/cloudquery/plugin-sdk/v4/types"
	resifaces "go.mws.cloud/go-sdk/pkg/resources/interfaces"
)

// MWS models are generated from OpenAPI and wrap plain values in types the
// default reflection-based transformer cannot see through:
//
//   - resource identifiers (AnyResourceID and the typed *ID types) keep the
//     path in an unexported field, so reflection yields an empty struct;
//   - references (*Ref) do the same, one level deeper;
//   - optional.Optional[T] and optional.OptionalNil[T] would each expand into
//     value/set/null columns instead of a single column of T.
//
// Left alone, a table gets no usable id column and therefore no primary key.

// resourceID is implemented by AnyResourceID and every generated *ID type.
// ID() is documented to be absolute and service-qualified, so it is the only
// form worth storing: "rm/projects/my-project".
type resourceID interface {
	ID() string
	ResourceName() resifaces.ResourceName
}

// resourceRef is implemented by every generated *Ref type. A reference keeps
// two forms of the same path and the difference matters here: Path() returns
// it as the server wrote it, which may be relative, while IDPath() is always
// absolute. See RefPath.
type resourceRef interface {
	IDPath() string
	Path() string
}

// quantity is implemented by the unit types the models measure things with:
// byte sizes, bitrates, frequencies, throughputs. Each keeps the amount and
// the unit it was written in, and BigInt returns the amount in the base unit
// of its dimension: bytes, bits per second, hertz. That is the form worth
// storing, because a disk of "10 GB" and one of "10240 MB" have to compare and
// add up in a query.
type quantity interface {
	BigInt() *big.Int
}

// cidrAddress and ipAddress are implemented by the address types, which the
// network models are full of. The version-specific variants, CIDR4Address and
// the like, embed the general type, so these two checks reach all of them.
type cidrAddress interface {
	ToNetCIDR() (net.IP, *net.IPNet)
}

type ipAddress interface {
	ToNetIP() net.IP
}

// redacted reports whether a field holds a secret: an api key, an hmac secret,
// the private half of a key pair.
//
// The SDK wraps those in a type that refuses to render itself, printing "***"
// as text, as json and in a log alike, and slog.LogValuer is the marker it
// carries, uniquely in the SDK. A column of "***" would only restate what the
// row already says, so such a field gets no column at all.
func redacted(t reflect.Type) bool {
	return implementsAny(t, logValuerType)
}

// rawJSON reports whether a byte slice holds json rather than bytes.
//
// A model that carries a document it does not describe itself, such as the
// settings of a database engine or the shape of a route, declares the field as
// json.RawMessage or as a named type over it, and those render themselves as
// json. A field that is genuinely binary, such as the CA certificate of a
// Kubernetes cluster, is a plain []byte and renders as base64. Both reflect as
// byte slices, and only the first is worth a queryable column.
func rawJSON(t reflect.Type) bool {
	return implementsAny(t, jsonMarshalerType)
}

// unwrapOptional reports whether t is optional.Optional[T] or
// optional.OptionalNil[T] and returns the type of its Value field.
func unwrapOptional(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() != reflect.Struct {
		return nil, false
	}
	value, ok := t.FieldByName("Value")
	if !ok {
		return nil, false
	}
	if _, ok := t.FieldByName("Set"); !ok {
		return nil, false
	}
	return value.Type, true
}

// implementsAny reports whether t or *t implements the given interface.
func implementsAny(t reflect.Type, iface reflect.Type) bool {
	return t.Implements(iface) || reflect.PointerTo(t).Implements(iface)
}

var (
	resourceIDType  = reflect.TypeOf((*resourceID)(nil)).Elem()
	resourceRefType = reflect.TypeOf((*resourceRef)(nil)).Elem()
	quantityType    = reflect.TypeOf((*quantity)(nil)).Elem()
	cidrAddressType = reflect.TypeOf((*cidrAddress)(nil)).Elem()
	ipAddressType   = reflect.TypeOf((*ipAddress)(nil)).Elem()
	stringerType    = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	logValuerType   = reflect.TypeOf((*slog.LogValuer)(nil)).Elem()

	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
)

// mwsTypeToArrow resolves a reflect.Type to an Arrow DataType, peeling
// pointers, optionals and slices on the way. It returns nil to let the default
// transformer decide.
func mwsTypeToArrow(t reflect.Type) (arrow.DataType, error) {
	switch t.Kind() {
	case reflect.Pointer:
		return mwsTypeToArrow(t.Elem())
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			if rawJSON(t) {
				return cqtypes.ExtensionTypes.JSON, nil
			}
			return nil, nil // []byte
		}
		elem, err := mwsTypeToArrow(t.Elem())
		if err != nil || elem == nil {
			return nil, err
		}
		if elem == cqtypes.ExtensionTypes.JSON {
			return cqtypes.ExtensionTypes.JSON, nil
		}
		return arrow.ListOf(elem), nil
	case reflect.Interface:
		return nil, nil
	}

	if inner, ok := unwrapOptional(t); ok {
		// The column has to describe what the optional holds. Returning nil
		// here would hand the caller back to the SDK default, which sees the
		// wrapper struct and settles for json.
		dataType, err := mwsTypeToArrow(inner)
		if err != nil || dataType != nil {
			return dataType, err
		}
		return transformers.DefaultTypeTransformer(reflect.StructField{Type: inner})
	}
	if implementsAny(t, resourceIDType) || implementsAny(t, resourceRefType) {
		return arrow.BinaryTypes.String, nil
	}
	// Quantities and addresses keep their state unexported too, so they would
	// otherwise be rendered as text by the rule below. They are asked first
	// because a number that adds up and an address the warehouse understands
	// are both worth more than the string they print as.
	if implementsAny(t, quantityType) {
		return arrow.PrimitiveTypes.Int64, nil
	}
	if implementsAny(t, cidrAddressType) || implementsAny(t, ipAddressType) {
		return cqtypes.ExtensionTypes.Inet, nil
	}
	if opaqueStringer(t) {
		return arrow.BinaryTypes.String, nil
	}
	return nil, nil
}

// opaqueStringer reports whether a type keeps everything unexported but can
// render itself, which is how the SDK models units: a duration holds its value
// and the raw "30s" it was parsed from, both unexported. Reflection sees an
// empty struct there, so without this the field would flatten into nothing.
func opaqueStringer(t reflect.Type) bool {
	// Named scalars, such as the string enums the models use for states, also
	// render themselves but need no special handling.
	if t.Kind() != reflect.Struct || !implementsAny(t, stringerType) {
		return false
	}
	// time.Time is the shape this describes exactly: unexported fields and a
	// String method. It must stay a timestamp, so rendering only takes over
	// where the SDK has nothing better than an opaque blob to offer.
	if dataType, err := transformers.DefaultTypeTransformer(reflect.StructField{Type: t}); err == nil &&
		dataType != nil && !arrow.TypeEqual(dataType, cqtypes.ExtensionTypes.JSON) {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).IsExported() {
			return false
		}
	}
	return true
}

// TypeTransformer maps MWS wrapper types onto Arrow types.
func TypeTransformer(field reflect.StructField) (arrow.DataType, error) {
	return mwsTypeToArrow(field.Type)
}

// mwsResolver returns the resolver able to read the value the Arrow type
// promises.
func mwsResolver(t reflect.Type, path string) schema.ColumnResolver {
	value := mwsConverter(t)
	if value == nil {
		return nil
	}
	return resolveValue(path, value)
}

// mwsConverter mirrors mwsTypeToArrow: that decides what a column promises,
// this delivers it. A nil converter means the SDK default already reads the
// field correctly, and the two have to be changed together.
func mwsConverter(t reflect.Type) converter {
	switch t.Kind() {
	case reflect.Pointer:
		return mwsConverter(t.Elem())
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return nil // []byte
		}
		elem := mwsConverter(t.Elem())
		if elem == nil {
			return nil
		}
		return sliceValue(elem)
	case reflect.Interface:
		return nil
	}

	if inner, ok := unwrapOptional(t); ok {
		return optionalValue(mwsConverter(inner))
	}
	if implementsAny(t, resourceIDType) {
		return resourceIDValue
	}
	if implementsAny(t, resourceRefType) {
		return refValue
	}
	if implementsAny(t, quantityType) {
		return quantityValue
	}
	if implementsAny(t, cidrAddressType) || implementsAny(t, ipAddressType) {
		return inetValue
	}
	if opaqueStringer(t) {
		return stringerValue
	}
	return nil
}

// ResolverTransformer returns a custom ColumnResolver for MWS struct fields.
// Paths are dotted and read with funk.Get at resolve time.
func ResolverTransformer(field reflect.StructField, path string) schema.ColumnResolver {
	return mwsResolver(field.Type, path)
}
