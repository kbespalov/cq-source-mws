# CloudQuery source for MWS Cloud Platform

[![CI](https://github.com/kbespalov/cq-source-mws/actions/workflows/ci.yml/badge.svg)](https://github.com/kbespalov/cq-source-mws/actions/workflows/ci.yml)
[![CodeQL](https://github.com/kbespalov/cq-source-mws/actions/workflows/codeql.yml/badge.svg)](https://github.com/kbespalov/cq-source-mws/actions/workflows/codeql.yml)

A [CloudQuery](https://www.cloudquery.io/) source plugin for the [MWS Cloud Platform](https://mws.ru/cloud-platform/). It lists the resources a set of credentials can see and writes them to any CloudQuery destination — PostgreSQL, SQLite, S3, and the rest of the [destination catalog](https://hub.cloudquery.io/plugins/destination).

Auth is the same as the MWS CLI and Terraform provider: an IAM token or a service-account key. The plugin discovers the organization / folder / project tree from the project list, then walks every selected project. A service that was never enabled in a project is skipped, not treated as a failed sync.

## Configuration

```yaml
kind: source
spec:
  name: mws
  registry: local
  path: ./cq-source-mws
  tables: ["*"]
  destinations: [sqlite]
  spec:
    # All three are optional. Empty means everything the credentials reach.
    # An installation has hundreds of projects; narrowing this list is the
    # main lever on sync duration.
    organization_ids: []
    folder_ids: []
    project_ids: []
    concurrency: 10
    scheduler: dfs          # dfs | round-robin | shuffle | shuffle-queue
    timeout: 30s
    debug: false            # log HTTP calls made by the SDK
---
kind: destination
spec:
  name: sqlite
  registry: local
  path: ./dist/cq-destination-sqlite
  spec:
    connection_string: ./test/output/mws.db
```

`test/config.yaml` is a ready-to-run copy of the above.

## Authentication

Set one of:

- `MWS_TOKEN` — an IAM token (`mws iam create-token`)
- `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` — path to a service-account key

A user CLI profile that opens a browser will not work unattended. For CI and scheduled syncs use a service-account key.

## Build and test

The MWS Go SDK needs Go 1.26 or newer. The toolchain is pulled in automatically.

```sh
make build          # ./cq-source-mws
make test           # go test -race ./...
make lint           # golangci-lint
make destination    # local SQLite destination, no CloudQuery Hub account
make sync           # needs MWS credentials; writes test/output/mws.db
```

Pull requests and pushes to `main` run lint, tests, `go mod tidy`, govulncheck, and a build. A tag `v*` publishes binaries on the GitHub release. Dependabot opens a weekly PR for Go modules and Actions.

If `go build` cannot download the 1.26 toolchain (a corporate proxy is the usual reason):

```sh
GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org make build
```

## Tables

Every listable resource in the [official Go SDK](https://github.com/mws-cloud-platform/go-sdk) is synced. Nested resources are child tables: a subnet hangs off its network, a Kafka topic off its cluster. Secret material (API keys, HMAC secrets, private keys, secret payloads) is never stored.

Service names in the table link to the overview in [MWS Docs](https://mws.ru/docs/). The full catalog of platform services and their launch stages is [here](https://mws.ru/docs/cloud-platform/about/general/launch-stages.html).

| Service | Resource | Table | Status |
|---|---|---|:---:|
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | [Project](https://mws.ru/docs/cloud-platform/org-rm/general/project-overview.html) | `mws_resmanager_projects` | ✅ |
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | Enabled service | `mws_resmanager_enabled_services` | ✅ |
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | Region | `mws_resmanager_regions` | ✅ |
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | Zone | `mws_resmanager_zones` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | Virtual machine | `mws_compute_virtual_machines` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | Disk | `mws_compute_disks` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | Snapshot | `mws_compute_snapshots` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | Image | `mws_compute_images` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | Disk backup | `mws_compute_disk_backups` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | VM type | `mws_compute_vm_types` | ✅ |
| [Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) | Disk type | `mws_compute_disk_types` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Network | `mws_vpc_networks` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Subnet | `mws_vpc_subnets` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Route | `mws_vpc_routes` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Firewall rule | `mws_vpc_firewall_rules` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Address | `mws_vpc_addresses` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Address group | `mws_vpc_address_groups` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | Egress NAT | `mws_vpc_egress_nats` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | One-to-one NAT | `mws_vpc_one_to_one_nats` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | NAT gateway | `mws_vpc_nat_gateways` | ✅ |
| [VPC](https://mws.ru/docs/cloud-platform/vpc/general/whatis-vpc.html) | External address | `mws_vpc_external_addresses` | ✅ |
| [Network Load Balancer](https://mws.ru/docs/cloud-platform/nlb/general/whatis-nlb.html) | Load balancer | `mws_nlb_load_balancers` | ✅ |
| [Managed PostgreSQL](https://mws.ru/docs/cloud-platform/mpostgres/general/whatis-mpostgres.html) | Cluster | `mws_mpostgres_clusters` | ✅ |
| [Managed PostgreSQL](https://mws.ru/docs/cloud-platform/mpostgres/general/whatis-mpostgres.html) | Cluster user | `mws_mpostgres_cluster_users` | ✅ |
| [Managed PostgreSQL](https://mws.ru/docs/cloud-platform/mpostgres/general/whatis-mpostgres.html) | Cluster database | `mws_mpostgres_cluster_databases` | ✅ |
| [Managed PostgreSQL](https://mws.ru/docs/cloud-platform/mpostgres/general/whatis-mpostgres.html) | Backup | `mws_mpostgres_backups` | ✅ |
| [Managed ClickHouse](https://mws.ru/docs/cloud-platform/mclickhouse/general/whatis-mclickhouse.html) | Cluster | `mws_mclickhouse_clusters` | ✅ |
| [Managed ClickHouse](https://mws.ru/docs/cloud-platform/mclickhouse/general/whatis-mclickhouse.html) | Cluster user | `mws_mclickhouse_cluster_users` | ✅ |
| [Managed ClickHouse](https://mws.ru/docs/cloud-platform/mclickhouse/general/whatis-mclickhouse.html) | Backup | `mws_mclickhouse_backups` | ✅ |
| [Managed Kafka](https://mws.ru/docs/cloud-platform/mkafka/general/whatis-mkafka.html) | Cluster | `mws_mkafka_clusters` | ✅ |
| [Managed Kafka](https://mws.ru/docs/cloud-platform/mkafka/general/whatis-mkafka.html) | Topic | `mws_mkafka_topics` | ✅ |
| [Managed Kafka](https://mws.ru/docs/cloud-platform/mkafka/general/whatis-mkafka.html) | User | `mws_mkafka_users` | ✅ |
| [Managed Kafka](https://mws.ru/docs/cloud-platform/mkafka/general/whatis-mkafka.html) | Connector | `mws_mkafka_connectors` | ✅ |
| [Managed Kubernetes](https://mws.ru/docs/cloud-platform/mk8s/general/whatis-mk8s.html) | Cluster | `mws_mk8s_clusters` | ✅ |
| [Managed Kubernetes](https://mws.ru/docs/cloud-platform/mk8s/general/whatis-mk8s.html) | Node group | `mws_mk8s_node_groups` | ✅ |
| [Managed Kubernetes](https://mws.ru/docs/cloud-platform/mk8s/general/whatis-mk8s.html) | Release channel | `mws_mk8s_release_channels` | ✅ |
| [Managed Kubernetes](https://mws.ru/docs/cloud-platform/mk8s/general/whatis-mk8s.html) | Version | `mws_mk8s_versions` | ✅ |
| [IAM](https://mws.ru/docs/cloud-platform/iam/general/whatis-iam.html) | Service account | `mws_iam_service_accounts` | ✅ |
| [IAM](https://mws.ru/docs/cloud-platform/iam/general/whatis-iam.html) | API key | `mws_iam_api_keys` | ✅ |
| [IAM](https://mws.ru/docs/cloud-platform/iam/general/whatis-iam.html) | Authorized key | `mws_iam_authorized_keys` | ✅ |
| [IAM](https://mws.ru/docs/cloud-platform/iam/general/whatis-iam.html) | HMAC key | `mws_iam_hmac_keys` | ✅ |
| [IAM](https://mws.ru/docs/cloud-platform/iam/general/whatis-iam.html) | Role | `mws_iam_global_roles` | ✅ |
| [KMS](https://mws.ru/docs/cloud-platform/kms/general/whatis-kms.html) | Crypto key | `mws_kms_crypto_keys` | ✅ |
| [KMS](https://mws.ru/docs/cloud-platform/kms/general/whatis-kms.html) | Crypto key version | `mws_kms_crypto_key_versions` | ✅ |
| [KMS](https://mws.ru/docs/cloud-platform/kms/general/whatis-kms.html) | Role binding | `mws_kms_crypto_key_role_bindings` | ✅ |
| [Secret Manager](https://mws.ru/docs/cloud-platform/secret-manager/general/whatis-secret-manager.html) | Secret | `mws_secretmanager_secrets` | ✅ |
| [Secret Manager](https://mws.ru/docs/cloud-platform/secret-manager/general/whatis-secret-manager.html) | Secret version | `mws_secretmanager_secret_versions` | ✅ |
| [Secret Manager](https://mws.ru/docs/cloud-platform/secret-manager/general/whatis-secret-manager.html) | Role binding | `mws_secretmanager_secret_role_bindings` | ✅ |
| [Certificate Manager](https://mws.ru/docs/cloud-platform/certmanager/general/whatis-cert-manager.html) | Certificate | `mws_certmanager_certificates` | ✅ |
| [Certificate Manager](https://mws.ru/docs/cloud-platform/certmanager/general/whatis-cert-manager.html) | Role binding | `mws_certmanager_certificate_role_bindings` | ✅ |
| [MWS GPT](https://mws.ru/docs/cloud-platform/gpt/general/whatis-gpt.html) | Model | `mws_gpt_models` | ✅ |
| [MWS GPT](https://mws.ru/docs/cloud-platform/gpt/general/whatis-gpt.html) | Deployment | `mws_gpt_deployments` | ✅ |

### Not yet supported

**Project services without a Go SDK client.** They have a REST API and share the same resource envelope as the tables above, so they can reuse `TransformResource` once a lister exists.

| Service | Status |
|---|:---:|
| [Object Storage](https://mws.ru/docs/cloud-platform/storage/general/whatis-object-storage.html) | ❌ |
| [Private DNS](https://mws.ru/docs/cloud-platform/vpc/general/vpc-private-dns-overview.html) | ❌ |
| [Artifact Registry](https://mws.ru/docs/cloud-platform/registry/general/whatis-registry.html) | ❌ |
| [CDN](https://mws.ru/docs/cloud-platform/cdn/general/whatis-cdn.html) | ❌ |
| [Audit Logs](https://mws.ru/docs/cloud-platform/audit-logs/general/whatis-audit-logs.html) | ❌ |
| Queues | ❌ |

**Organization and identity.** There is no public list API in the Go SDK for the organization itself. The plugin reconstructs organization and folder IDs from the project list; empty folders, members, and federations are not synced. Role bindings exist only on KMS keys, secrets, and certificates — not on the organization, folder, or project.

| Service | Resource | Status |
|---|---|:---:|
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | [Organization](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-organization.html) | ❌ |
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | [Folder](https://mws.ru/docs/cloud-platform/org-rm/general/folder-overview.html) | ❌ |
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | [Federation](https://mws.ru/docs/cloud-platform/org-rm/general/federation-overview.html) | ❌ |
| [Resource Manager](https://mws.ru/docs/cloud-platform/org-rm/general/whatis-rm.html) | [User](https://mws.ru/docs/cloud-platform/org-rm/general/user-in-org-operations.html) | ❌ |
| [IAM](https://mws.ru/docs/cloud-platform/iam/general/whatis-iam.html) | Access binding (organization / folder / project) | ❌ |

**Billing and other platform surfaces.** These are not listable inventory in the Go SDK.

| Service | Status |
|---|:---:|
| [Billing](https://mws.ru/docs/cloud-platform/billing/general/whatis-billing.html) | ❌ |
| [Monitoring](https://mws.ru/docs/cloud-platform/monitoring/general/whatis-monitoring.html) | ❌ |

See [ARCHITECTURE.md](ARCHITECTURE.md) for how models become columns and how the project hierarchy is discovered.
