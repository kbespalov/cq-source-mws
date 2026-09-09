package client

import (
	"fmt"
	"reflect"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	cqtypes "github.com/cloudquery/plugin-sdk/v4/types"
)

// Every MWS resource is shaped like a Kubernetes object: a kind, a metadata
// block that is identical across services, a spec holding the desired state and
// a status holding the observed one. The generated Go models mirror that, and
// they nest freely: a virtual machine keeps its hostname under spec.os and its
// core count under spec.hardware.power.
//
// The SDK struct transformer only unwraps one level, so anything deeper becomes
// an opaque json column. TransformResource walks the model itself and applies a
// single rule: a field that the type system can express as a real column
// becomes one, a struct that would otherwise be an opaque blob is flattened
// into its fields, and only collections stay json. Collections are where child
// tables belong, not columns.
//
// Metadata fields become top-level columns because they are the same for every
// resource and read like the identity columns other CloudQuery sources expose.
// Everything else keeps its prefix, so a spec field cannot collide with a
// status field of the same name and the origin of a column stays readable.
const (
	metadataField = "Metadata"

	// idColumn is the flattened name of metadata.id and the primary key of
	// every resource table.
	idColumn = "id"
	// nameColumn holds the resource name on its own. An id is an absolute
	// path, "compute/projects/my-project/virtualMachines/my-vm", which is what
	// joins need but not what anyone wants to read; the SDK derives the
	// trailing name from the same identifier. Services differ in whether their
	// metadata states the name as well, so it is only derived where it is
	// missing, and the column means the same thing either way.
	nameColumn = "name"
	nameField  = "Name"

	// maxDepth bounds the walk. Nothing in the models comes close, but a
	// self-referential type would otherwise recurse forever.
	maxDepth = 6
)

// TransformResource builds the column set of a table from an MWS model.
func TransformResource(model any) schema.Transform {
	return func(table *schema.Table) error {
		t := reflect.TypeOf(model)
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return fmt.Errorf("table %q: expected a struct model, got %s", table.Name, t.Kind())
		}

		columns, err := flatten(t, "", "", 0)
		if err != nil {
			return fmt.Errorf("table %q: %w", table.Name, err)
		}

		id := false
		seen := make(map[string]struct{}, len(columns))
		for i, col := range columns {
			if _, duplicate := seen[col.Name]; duplicate {
				return fmt.Errorf("table %q: model %s produces two %q columns", table.Name, t.Name(), col.Name)
			}
			seen[col.Name] = struct{}{}
			if col.Name == idColumn {
				columns[i].PrimaryKey = true
				columns[i].NotNull = true
				id = true
			}
		}
		if !id {
			return fmt.Errorf("table %q: model %s has no metadata.id to use as primary key", table.Name, t.Name())
		}

		table.Columns = append(table.Columns, columns...)
		return nil
	}
}

// flatten walks a struct type and returns one column per leaf field.
//
// namePrefix accumulates the column name ("spec_os_"), path accumulates the
// funk.Get path used at resolve time ("Spec.Os.Value."). Embedded structs
// extend the path but not the name, which is what makes the shared metadata
// fields surface as if they were declared on the resource itself.
func flatten(t reflect.Type, namePrefix, path string, depth int) ([]schema.Column, error) {
	var columns []schema.Column

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() || redacted(field.Type) {
			continue
		}

		dataType, err := columnType(field)
		if err != nil {
			return nil, err
		}

		nested, suffix, canDescend := walkable(field.Type)
		if !canDescend || depth >= maxDepth || !opaque(dataType) {
			col, err := column(field, dataType, namePrefix, path)
			if err != nil {
				return nil, err
			}
			if col == nil {
				continue
			}
			columns = append(columns, *col)
			if _, named := t.FieldByName(nameField); col.Name == idColumn && !named {
				columns = append(columns, schema.Column{
					Name:     nameColumn,
					Type:     arrow.BinaryTypes.String,
					Resolver: ResolveResourceName(path + field.Name),
				})
			}
			continue
		}

		prefix := namePrefix
		if !field.Anonymous && (depth != 0 || field.Name != metadataField) {
			name, err := transformers.DefaultNameTransformer(field)
			if err != nil {
				return nil, err
			}
			prefix += name + "_"
		}

		cols, err := flatten(nested, prefix, path+field.Name+"."+suffix, depth+1)
		if err != nil {
			return nil, err
		}
		if len(cols) == 0 {
			// The struct had nothing reflection could reach. Keeping the field
			// as json is worse than columns but far better than the field
			// quietly disappearing from the table.
			col, err := column(field, dataType, namePrefix, path)
			if err != nil {
				return nil, err
			}
			if col != nil {
				columns = append(columns, *col)
			}
			continue
		}
		columns = append(columns, cols...)
	}

	return columns, nil
}

// columnType is the type a field would get as a column: the MWS-aware mapping
// first, the SDK default second.
func columnType(field reflect.StructField) (arrow.DataType, error) {
	dataType, err := TypeTransformer(field)
	if err != nil || dataType != nil {
		return dataType, err
	}
	return transformers.DefaultTypeTransformer(field)
}

// opaque reports whether a type carries no queryable structure, which is the
// signal that the field is worth flattening instead.
func opaque(dataType arrow.DataType) bool {
	return dataType == nil || arrow.TypeEqual(dataType, cqtypes.ExtensionTypes.JSON)
}

func column(field reflect.StructField, dataType arrow.DataType, namePrefix, path string) (*schema.Column, error) {
	if dataType == nil {
		return nil, nil // the SDK ignores this field, so do we
	}

	name, err := transformers.DefaultNameTransformer(field)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, nil
	}

	fieldPath := path + field.Name
	resolver := ResolverTransformer(field, fieldPath)
	if resolver == nil {
		resolver = transformers.DefaultResolverTransformer(field, fieldPath)
	}

	return &schema.Column{
		Name:     namePrefix + name,
		Type:     dataType,
		Resolver: resolver,
	}, nil
}

// walkable reports whether a field holds a struct worth descending into, and
// returns that struct along with the path segments reaching it.
//
// Services differ in how they wrap the same block: a project carries metadata
// as a plain pointer, a virtual machine wraps it in optional.OptionalNil. Both
// have to flatten the same way, so pointers and optionals are peeled here and
// the optional's "Value" hop is folded into the path.
func walkable(t reflect.Type) (reflect.Type, string, bool) {
	suffix := ""
	for {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return nil, "", false
		}
		inner, ok := unwrapOptional(t)
		if !ok {
			return t, suffix, true
		}
		t = inner
		suffix += "Value."
	}
}
