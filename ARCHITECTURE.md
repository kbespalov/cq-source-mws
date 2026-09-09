# Architecture

This plugin is a CloudQuery source. The interesting part is not the CloudQuery
shell — that is a thin `serve.Plugin` around a scheduler — but how MWS models
become columns, and how one list of projects becomes the multiplexer for
everything else.

## Resource shape

Every MWS resource is a Kubernetes-style envelope: `kind`, `metadata`, `spec`,
`status`. `metadata.id` is the primary key and an absolute path
(`rm/projects/my-project`, `compute/projects/my-project/virtualMachines/vm-1`).

The generated Go models wrap almost every interesting field:

- identifiers and references keep the path in an unexported field
- quantities (`ByteSize`, `Throughput`, …) keep amount and unit unexported
- addresses (`IPAddress`, `CIDRAddress`) do the same
- `optional.Optional[T]` would otherwise expand into `value` / `set` / `null`

`client.TransformResource` walks the struct, hoists metadata to top-level
columns, prefixes everything else with `spec_` / `status_`, and maps those
wrappers onto Arrow types. Slices of structs stay JSON: that is where a child
table belongs, not a column. Fields the SDK marks as secrets
(`sensitive.Sensitive`, detected via `slog.LogValuer`) get no column at all.

## Hierarchy

MWS has no public list API for organizations or folders. Every project reports
both of its ancestors, so one `ListProjectsV3` call reconstructs the tree.
`organization_ids`, `folder_ids` and `project_ids` in the spec filter that
tree. Multiplexers then clone the client once per organization, folder or
project.

A project that never enabled a service answers 403 or 404. `Client.Skip`
treats those as empty rather than fatal, otherwise a full sync would never
finish. The cost is that a missing IAM role looks the same as an unused
service; every skip is logged.

## Adding a table

A listable SDK type is one file under `resources/<service>/`:

```go
func Disks() *schema.Table {
    return &schema.Table{
        Name:        "mws_compute_disks",
        Description: "Disks of every synced project",
        Resolver:    fetchDisks,
        Multiplex:   client.ProjectMultiplex(),
        Transform:   client.TransformResource(&computemodel.DiskOptionalResponse{}),
        Columns:     client.ProjectHierarchyColumns(),
    }
}
```

Register it in `plugin/tables.go`. Nested listings take the parent name from
`client.ParentResourceName` and use `client.ChildColumns("cluster_id")` (or
`network_id`, `secret_id`, …). Listings without a page token use
`client.Fetch` instead of `client.List`.

A table whose model has no `metadata.id` — today only Kubernetes release
channels and versions — spells its columns out by hand and puts `project_id`
in the primary key.

## What this plugin does not cover

Services that have a REST API but no Go SDK client yet: object storage,
private DNS, container registry, CDN, audit logs, queues. They share the
same envelope and can reuse `TransformResource` once a lister exists.

Organization and identity live above the project list: organizations,
folders, federations, users, and access bindings on the organization /
folder / project. Those IDs are inferred from projects; there is no
public list API in the SDK. Billing accounts and monitoring metrics are
the same kind of gap — a platform surface without a listable client.

Endpoints that only `Get` a single object, or that perform an action
(switchover, decrypt, issue a token), are not tables.
