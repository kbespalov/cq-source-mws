# Contributing

## Layout

- `client/` — SDK session, hierarchy, multiplexers, and the type walk that
  builds columns. This is where a new wrapper type is taught to the plugin.
- `resources/<service>/` — one file per table, all the same shape.
- `plugin/` — CloudQuery glue: spec JSON, scheduler, table registry.

The current design is in [ARCHITECTURE.md](ARCHITECTURE.md).

## Tests

`make test` is the default gate. `client` tests pin the type walk against
fixtures so a change to flattening cannot silently turn an address into text
or a secret into a column. `plugin` tests assert every table has a primary
key, a multiplexer (top-level only) and a unique `mws_` name.

`TestSyncProjects` talks to a live installation and is skipped unless
`MWS_TOKEN` or `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` is set.

## Adding a service

1. Confirm the SDK has a `List*` method. A `Get*` alone is not a table.
2. Copy the nearest existing table. Project-scoped listings use
   `client.List` and `c.Skip(serviceName)`. Nested listings take the parent
   name from `client.ParentResourceName`.
3. Register the root table in `plugin/tables.go`. Children go in `Relations`.
4. Run `go test ./plugin/ -run TestSchemas` — it walks `FlattenTables()`, so
   a nested table is checked automatically.

Do not store secret material. If a field implements `slog.LogValuer`, the
transformer already drops it; a new secret type that does not should be
taught there, not special-cased in the table.
