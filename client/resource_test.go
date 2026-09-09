package client

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/cloudquery/plugin-sdk/v4/scalar"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	cqtypes "github.com/cloudquery/plugin-sdk/v4/types"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	clickhousemodel "go.mws.cloud/go-sdk/service/mclickhouse/model"
	mk8smodel "go.mws.cloud/go-sdk/service/mk8s/model"
	rmmodel "go.mws.cloud/go-sdk/service/rm/model"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
)

// resolve builds the table for a model, fills it from payload and returns the
// resolved column values by name. Columns whose resolver needs a multiplexed
// client, such as the hierarchy ones, are not part of the transform and so are
// not exercised here.
func resolve(t *testing.T, model any, payload string) map[string]string {
	t.Helper()
	return resolveMeta(t, nil, model, payload)
}

func resolveMeta(t *testing.T, meta schema.ClientMeta, model any, payload string) map[string]string {
	t.Helper()

	if err := json.Unmarshal([]byte(payload), model); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	table := &schema.Table{Name: "test"}
	if err := TransformResource(model)(table); err != nil {
		t.Fatalf("transform: %v", err)
	}

	resource := schema.NewResourceData(table, nil, model)
	values := make(map[string]string, len(table.Columns))
	for _, col := range table.Columns {
		if err := col.Resolver(context.Background(), meta, resource, col); err != nil {
			t.Fatalf("resolve %q: %v", col.Name, err)
		}
		values[col.Name] = resource.Get(col.Name).String()
	}
	return values
}

// TestReferencesAreAbsolute pins down the normalisation of references.
//
// The API writes the same kind of path in two formats: this is a real project
// response, where the folder came absolute and the organization relative.
// Storing them as received would leave two spellings of one organization in the
// warehouse, and no join between them.
func TestReferencesAreAbsolute(t *testing.T) {
	const payload = `{
	  "kind": "Project",
	  "metadata": {"id": "rm/projects/my-project", "displayName": "my-project"},
	  "spec": {"folder": "org/organizations/my-org/folders/my-folder"},
	  "status": {"organization": "organizations/my-org", "active": true}
	}`

	values := resolve(t, &rmmodel.ProjectResponse{}, payload)

	for column, want := range map[string]string{
		"id":                  "rm/projects/my-project",
		"name":                "my-project",
		"display_name":        "my-project",
		"spec_folder":         "org/organizations/my-org/folders/my-folder",
		"status_organization": "org/organizations/my-org",
		"status_active":       "true",
	} {
		if got := values[column]; got != want {
			t.Errorf("column %q: got %q, want %q", column, got, want)
		}
	}

	// Two tables write their own identifier as a reference rather than as a
	// resource id, so the name has to be derived from the path either way.
	values = resolve(t, &rmmodel.RegionResponse{}, `{
	  "kind": "Region",
	  "metadata": {"id": "rm/regions/ru-central1"},
	  "spec": {}
	}`)
	for column, want := range map[string]string{
		"id":   "rm/regions/ru-central1",
		"name": "ru-central1",
	} {
		if got := values[column]; got != want {
			t.Errorf("column %q: got %q, want %q", column, got, want)
		}
	}

	// A NIC list stays JSON. The server writes relative refs; the column
	// stores the same absolute paths the address and NAT tables use as id.
	t.Run("json nics", func(t *testing.T) {
		const vm = `{
		  "kind": "VirtualMachine",
		  "metadata": {"id": "compute/projects/my-project/virtualMachines/vm-1"},
		  "spec": {"zone": "ru-central-1a"},
		  "status": {
		    "network": {
		      "networkInterfaces": [{
		        "name": "nic0",
		        "addresses": [{
		          "subnet": "projects/my-project/networks/net-1/subnets/subnet-1",
		          "network": "projects/my-project/networks/net-1",
		          "ref": "projects/my-project/networks/net-1/addresses/addr-1",
		          "ipAddress": "10.0.0.5",
		          "oneToOneNat": {
		            "ref": "projects/my-project/networks/net-1/oneToOneNats/nat-1",
		            "external": {
		              "ref": "projects/my-project/externalAddresses/ext-1",
		              "ipAddress": "203.0.113.10"
		            }
		          }
		        }]
		      }]
		    }
		  }
		}`

		rec := record(t, &computemodel.VirtualMachineOptionalResponse{}, vm)
		got := columnValue(t, rec, "status_network_network_interfaces")
		for _, want := range []string{
			"vpc/projects/my-project/networks/net-1/addresses/addr-1",
			"vpc/projects/my-project/networks/net-1/subnets/subnet-1",
			"vpc/projects/my-project/networks/net-1",
			"vpc/projects/my-project/networks/net-1/oneToOneNats/nat-1",
			"vpc/projects/my-project/externalAddresses/ext-1",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("NIC JSON missing %q\n%s", want, got)
			}
		}
		if strings.Contains(got, `"ref":"projects/`) {
			t.Errorf("NIC JSON still has a relative ref\n%s", got)
		}
	})

	// A disk named only as "disk-1" is completed from the multiplex project.
	t.Run("short disk name", func(t *testing.T) {
		const vm = `{
		  "kind": "VirtualMachine",
		  "metadata": {"id": "compute/projects/my-project/virtualMachines/vm-1"},
		  "spec": {
		    "zone": "ru-central-1a",
		    "storage": {"disks": [{"disk": {"ref": "disk-1"}}]}
		  },
		  "status": {}
		}`

		rec := recordMeta(t, &Client{ProjectName: "my-project"}, &computemodel.VirtualMachineOptionalResponse{}, vm)
		got := columnValue(t, rec, "spec_storage_disks")
		if !strings.Contains(got, "compute/projects/my-project/disks/disk-1") {
			t.Errorf("short disk name was not expanded\n%s", got)
		}
	})
}

// TestTimestamps covers time, which shows up in nearly every table.
//
// The models declare it two ways, *time.Time and optional.Optional[time.Time],
// and the SDK parses both with time.Parse(time.RFC3339, ...), which keeps
// whatever offset the string carried. Two things follow and are worth pinning
// down: the column has to stay a timestamp rather than text, and an offset has
// to end up as the same instant in UTC rather than a shifted wall clock.
func TestTimestamps(t *testing.T) {
	t.Run("column type", func(t *testing.T) {
		for _, tc := range []struct {
			model  any
			column string
		}{
			{&rmmodel.ProjectResponse{}, "create_time"},                         // *time.Time
			{&iammodel.AuthorizedKeyOptionalResponse{}, "spec_expiration_time"}, // optional.Optional[time.Time]
		} {
			table := &schema.Table{Name: "test"}
			if err := TransformResource(tc.model)(table); err != nil {
				t.Fatalf("transform: %v", err)
			}
			col := table.Column(tc.column)
			if col == nil {
				t.Fatalf("column %q is missing", tc.column)
			}
			if !arrow.TypeEqual(col.Type, arrow.FixedWidthTypes.Timestamp_us) {
				t.Errorf("column %q: got %s, want %s", tc.column, col.Type, arrow.FixedWidthTypes.Timestamp_us)
			}
		}
	})

	// The values are read back out of the Arrow record, which is what reaches
	// the destination, rather than off the scalar. The scalar's own String
	// formats with time.RFC3339 and hides fractional seconds, so asserting on
	// it would prove nothing about what is actually stored.
	t.Run("values", func(t *testing.T) {
		const payload = `{
		  "metadata": {
		    "id": "rm/projects/my-project",
		    "createTime": "2025-08-11T09:05:52.00217Z",
		    "updateTime": "2025-08-11T12:05:52+03:00",
		    "purgeTime": "2025-08-11T09:05:52.123456789Z"
		  },
		  "spec": {"folder": "org/organizations/o/folders/f"}
		}`

		record := record(t, &rmmodel.ProjectResponse{}, payload)

		for column, want := range map[string]string{
			// Sub-second precision survives.
			"create_time": "2025-08-11T09:05:52.00217Z",
			// Written with an offset, stored as the same instant in UTC.
			"update_time": "2025-08-11T09:05:52Z",
			// The column holds microseconds, so anything finer is truncated.
			"purge_time": "2025-08-11T09:05:52.123456Z",
			// Absent from the response: null, not the zero time.
			"delete_time": "(null)",
		} {
			if got := columnValue(t, record, column); got != want {
				t.Errorf("column %q: got %q, want %q", column, got, want)
			}
		}
	})
}

// TestQuantities covers the unit types the models measure hardware with.
//
// A disk size arrives as "10 GB" and the SDK keeps the amount and the unit
// apart, so the field renders itself but reflects as an empty struct. Rendered
// as text no query could add two disks up, let alone compare a size written in
// GB with one written in MB, so the column holds the amount in bytes, which is
// the form the SDK itself normalises to.
func TestQuantities(t *testing.T) {
	const payload = `{
	  "kind": "Disk",
	  "metadata": {"id": "compute/projects/my-project/disks/disk-1"},
	  "spec": {"zone": "ru-central-1a", "size": "10 GB"},
	  "status": {"size": "20 GB", "blockSize": "4 KB", "throughput": "450 MBps"}
	}`

	rec := record(t, &computemodel.DiskOptionalResponse{}, payload)

	if got := arrowType(t, &computemodel.DiskOptionalResponse{}, "spec_size"); !arrow.TypeEqual(got, arrow.PrimitiveTypes.Int64) {
		t.Errorf("column %q: got %s, want %s", "spec_size", got, arrow.PrimitiveTypes.Int64)
	}

	for column, want := range map[string]string{
		// optional.Optional[bytesize.ByteSize], 10 * 2^30.
		"spec_size": "10737418240",
		// *bytesize.ByteSize, the same treatment through a pointer.
		"status_size": "21474836480",
		// A unit small enough to read at a glance: 4 * 2^10.
		"status_block_size": "4096",
		// Rates are quantities too, in their own base unit of bytes per
		// second: 450 * 2^20.
		"status_throughput": "471859200",
		// Not in the response, and zero bytes would be a lie.
		"spec_block_size": "(null)",
	} {
		if got := columnValue(t, rec, column); got != want {
			t.Errorf("column %q: got %q, want %q", column, got, want)
		}
	}
}

// TestAddresses covers the address types, which every network table is built
// out of.
//
// Storing them as text would work, but the warehouse has a type for addresses
// that sorts and matches by subnet, so the column uses it. Two details have to
// survive the conversion: the host bits of a CIDR, which say which address in
// the subnet this is, and a list of addresses, which has to hold the same
// values a single address column would.
func TestAddresses(t *testing.T) {
	const subnet = `{
	  "kind": "Subnet",
	  "metadata": {"id": "vpc/projects/my-project/networks/net-1/subnets/subnet-1"},
	  "spec": {
	    "cidr": "10.0.1.5/24",
	    "dhcpOptions": {"domainNameServers": ["10.0.1.2", "8.8.8.8"]}
	  },
	  "status": {}
	}`

	if got := arrowType(t, &vpcmodel.SubnetOptionalResponse{}, "spec_cidr"); !arrow.TypeEqual(got, cqtypes.ExtensionTypes.Inet) {
		t.Errorf("column %q: got %s, want %s", "spec_cidr", got, cqtypes.ExtensionTypes.Inet)
	}

	rec := record(t, &vpcmodel.SubnetOptionalResponse{}, subnet)
	for column, want := range map[string]string{
		// The mask says /24 but the address is .5, and both are kept: the
		// subnet mask alone would lose which address this is.
		"spec_cidr": "10.0.1.5/24",
		// A list of addresses, each converted the way a single one would be.
		"spec_dhcp_options_domain_name_servers": `["10.0.1.2/32","8.8.8.8/32"]`,
		"spec_dhcp_options_ntp_servers":         "(null)",
	} {
		if got := columnValue(t, rec, column); got != want {
			t.Errorf("column %q: got %q, want %q", column, got, want)
		}
	}

	// A bare address reaches the same column type through a different model.
	const address = `{
	  "kind": "Address",
	  "metadata": {"id": "vpc/projects/my-project/networks/net-1/addresses/addr-1"},
	  "spec": {},
	  "status": {"ipAddress": "10.0.1.5"}
	}`

	rec = record(t, &vpcmodel.AddressOptionalResponse{}, address)
	if got, want := columnValue(t, rec, "status_ip_address"), "10.0.1.5/32"; got != want {
		t.Errorf("column %q: got %q, want %q", "status_ip_address", got, want)
	}
}

// TestSecretsGetNoColumn covers the fields the SDK refuses to render.
//
// An api key, an hmac secret and the private half of a key pair are wrapped in
// a type that prints "***" wherever it goes. Left to the ordinary rules it
// looks like any other type that keeps its state unexported, so it would get a
// column full of asterisks. The row already says the key exists, so the field
// is dropped instead, while the rest of the same block stays.
func TestSecretsGetNoColumn(t *testing.T) {
	for _, tc := range []struct {
		model   any
		dropped []string
		kept    string
	}{
		{&iammodel.ApiKeyResponse{}, []string{"status_api_key"}, "status_last_auth_time"},
		{&iammodel.HmacKeyResponse{}, []string{"status_secret_access_key"}, "status_access_key_id"},
		{
			&iammodel.AuthorizedKeyOptionalResponse{},
			[]string{"status_private_key", "status_private_key_file"},
			"spec_expiration_time",
		},
	} {
		table := &schema.Table{Name: "test"}
		if err := TransformResource(tc.model)(table); err != nil {
			t.Fatalf("transform: %v", err)
		}
		for _, column := range tc.dropped {
			if table.Column(column) != nil {
				t.Errorf("%T: column %q holds a secret and should not exist", tc.model, column)
			}
		}
		if table.Column(tc.kept) == nil {
			t.Errorf("%T: column %q went missing along with the secret", tc.model, tc.kept)
		}
	}
}

// TestRawJSON covers the fields a model declares as raw json because it does
// not model their contents: the settings of a database engine, the shape of a
// route that leaves the network locally.
//
// They reflect as byte slices, which the SDK default would store as a binary
// blob no query can reach into, so the column has to be json instead.
func TestRawJSON(t *testing.T) {
	const route = `{
	  "kind": "Route",
	  "metadata": {"id": "vpc/projects/my-project/networks/net-1/routes/route-1"},
	  "spec": {
	    "destination": {"cidr": "10.0.0.0/8"},
	    "nextHop": {"networkLocal": {}}
	  },
	  "status": {}
	}`

	if got := arrowType(t, &vpcmodel.RouteOptionalResponse{}, "spec_next_hop_network_local"); !arrow.TypeEqual(got, cqtypes.ExtensionTypes.JSON) {
		t.Errorf("column %q: got %s, want %s", "spec_next_hop_network_local", got, cqtypes.ExtensionTypes.JSON)
	}

	rec := record(t, &vpcmodel.RouteOptionalResponse{}, route)
	if got, want := columnValue(t, rec, "spec_next_hop_network_local"), "{}"; got != want {
		t.Errorf("column %q: got %q, want %q", "spec_next_hop_network_local", got, want)
	}

	// The same thing one wrapper deeper: a map of raw messages inside an
	// optional, which is how a cluster carries the settings of its engine.
	const cluster = `{
	  "kind": "ClickhouseCluster",
	  "metadata": {"id": "mclickhouse/projects/my-project/clusters/ch-1"},
	  "spec": {"version": "24.8", "config": {"max_connections": 4096}},
	  "status": {}
	}`

	rec = record(t, &clickhousemodel.ClickhouseClusterOptionalResponse{}, cluster)
	if got, want := columnValue(t, rec, "spec_config"), `{"max_connections":4096}`; got != want {
		t.Errorf("column %q: got %q, want %q", "spec_config", got, want)
	}

	// A region declares its spec as a named type over json.RawMessage, which
	// is the same thing wearing a different name.
	const region = `{
	  "kind": "Region",
	  "metadata": {"id": "rm/regions/ru-central1"},
	  "spec": {"zones": 3}
	}`

	rec = record(t, &rmmodel.RegionResponse{}, region)
	if got, want := columnValue(t, rec, "spec"), `{"zones":3}`; got != want {
		t.Errorf("column %q: got %q, want %q", "spec", got, want)
	}

	// A field that is genuinely binary keeps its bytes: the CA certificate of
	// a Kubernetes cluster is not json and must not be mistaken for it.
	if got := arrowType(t, &mk8smodel.ClusterOptionalResponse{}, "status_cluster_ca_certificate"); !arrow.TypeEqual(got, arrow.BinaryTypes.Binary) {
		t.Errorf("column %q: got %s, want %s", "status_cluster_ca_certificate", got, arrow.BinaryTypes.Binary)
	}
}

// arrowType returns the type the model gives a column.
func arrowType(t *testing.T, model any, column string) arrow.DataType {
	t.Helper()

	table := &schema.Table{Name: "test"}
	if err := TransformResource(model)(table); err != nil {
		t.Fatalf("transform: %v", err)
	}
	col := table.Column(column)
	if col == nil {
		t.Fatalf("column %q is missing", column)
	}
	return col.Type
}

// record builds the Arrow record a sync would emit for one resource.
func record(t *testing.T, model any, payload string) arrow.RecordBatch {
	t.Helper()
	return recordMeta(t, nil, model, payload)
}

func recordMeta(t *testing.T, meta schema.ClientMeta, model any, payload string) arrow.RecordBatch {
	t.Helper()

	if err := json.Unmarshal([]byte(payload), model); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	table := &schema.Table{Name: "test"}
	if err := TransformResource(model)(table); err != nil {
		t.Fatalf("transform: %v", err)
	}

	resource := schema.NewResourceData(table, nil, model)
	for _, col := range table.Columns {
		if err := col.Resolver(context.Background(), meta, resource, col); err != nil {
			t.Fatalf("resolve %q: %v", col.Name, err)
		}
	}

	builder := array.NewRecordBuilder(memory.DefaultAllocator, table.ToArrowSchema())
	for i, value := range resource.GetValues() {
		scalar.AppendToBuilder(builder.Field(i), value)
	}
	return builder.NewRecordBatch()
}

func columnValue(t *testing.T, rec arrow.RecordBatch, column string) string {
	t.Helper()

	indices := rec.Schema().FieldIndices(column)
	if len(indices) == 0 {
		t.Fatalf("column %q is missing from the record", column)
	}
	col := rec.Column(indices[0])
	if col.IsNull(0) {
		return "(null)"
	}
	return col.ValueStr(0)
}

// TestOptionalsUnwrap covers the other wrapper the models use.
//
// A virtual machine wraps metadata in optional.OptionalNil and its fields in
// optional.Optional. Both must yield the value they hold, and an absent one
// must be null rather than an empty string or a json blob.
func TestOptionalsUnwrap(t *testing.T) {
	const payload = `{
	  "kind": "VirtualMachine",
	  "metadata": {"id": "compute/projects/my-project/virtualMachines/vm-1", "displayName": "vm-1"},
	  "spec": {
	    "zone": "ru-central-1a",
	    "hardware": {"power": "ON", "gracefulShutdownTimeout": "30s"},
	    "os": {"hostname": "vm-1"}
	  },
	  "status": {"ready": {"state": "OK"}}
	}`

	values := resolve(t, &computemodel.VirtualMachineOptionalResponse{}, payload)

	for column, want := range map[string]string{
		"id":                 "compute/projects/my-project/virtualMachines/vm-1",
		"name":               "vm-1",
		"display_name":       "vm-1",
		"spec_zone":          "ru-central-1a",
		"spec_os_hostname":   "vm-1",
		"status_ready_state": "OK",
		// Nested blocks flatten instead of collapsing into json, and a unit
		// type that keeps its value unexported still renders itself.
		"spec_hardware_power": "ON",
		// Written as "30s", stored in the SDK's canonical rendering so the
		// same duration always reads the same way.
		"spec_hardware_graceful_shutdown_timeout": "PT30S",
		// description was not in the response: unset, not empty.
		"description": "(null)",
	} {
		if got := values[column]; got != want {
			t.Errorf("column %q: got %q, want %q", column, got, want)
		}
	}
}
