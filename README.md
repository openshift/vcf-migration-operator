# vcf-migration-operator

Kubernetes operator and OpenShift Console plugin for orchestrating migration of OpenShift clusters between VMware vCenters (e.g. VMware Cloud Foundation / VCF). Use the operator to drive the migration lifecycle and the console plugin to create migrations, browse vCenter inventory, and monitor progress from the OpenShift web console.

## Description

vcf-migration-operator automates moving an OpenShift cluster from a source vCenter to a target vCenter. The project consists of:

- **Operator**: A Kubebuilder-based controller that reconciles the `VmwareCloudFoundationMigration` custom resource. It prepares infrastructure (credentials, failure domains), initializes the destination, configures multi-site, migrates workload (machines/nodes), and cleans up the source. The operator uses the cluster's Machine API and OpenShift-specific resources; vSphere operations are performed via govmomi against the target vCenter.

- **OpenShift Console Plugin**: A dynamic plugin that adds a "VCF Migration" section to the OpenShift console (Administrator perspective). It provides a list of migrations, a create wizard (credentials + vCenter browse for failure domains), and a detail view with condition progress, live event stream (SSE), and machine topology. The plugin backend is a Go HTTP server that serves the webpack-built frontend and exposes API routes for vCenter inventory browsing and event streaming; the UI uses the console SDK for all Kubernetes CRUD and watch operations.

## Getting Started

### Prerequisites
- go version v1.25.0+
- podman
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make operator-image operator-push IMG=<some-registry>/vcf-migration-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don't work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/vcf-migration-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**

Apply the sample migration CRs from `config/samples/` (if present):

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## OpenShift Console Plugin

The console plugin runs as a separate Deployment and registers with the OpenShift console via a `ConsolePlugin` custom resource. It serves the UI and backend APIs (vCenter browse, event stream) over HTTPS.

### Prerequisites (plugin)

- Node.js 18+ and npm (for building the frontend)
- Go 1.25+ (backend shares the repo's Go module and reuses `internal/vsphere`)

### Build the plugin

```sh
# Frontend (webpack bundle into console-plugin/web/dist)
make console-plugin-frontend

# Backend (Go binary to bin/console-plugin)
make console-plugin-backend

# Container image (default tag: vcf-migration-console-plugin:latest)
make console-plugin-image
# Or with a custom tag:
make console-plugin-image CONSOLE_PLUGIN_IMG=<registry>/vcf-migration-console-plugin:<tag>
```

### Push the plugin image

```sh
make console-plugin-push CONSOLE_PLUGIN_IMG=<registry>/vcf-migration-console-plugin:<tag>
```

### Deploy the plugin

Ensure the plugin image is available to the cluster (e.g. push to the cluster's registry or load into Kind). Then:

```sh
make deploy-console-plugin CONSOLE_PLUGIN_IMG=<registry>/vcf-migration-console-plugin:<tag>
```

This applies the Kustomize manifests under `console-plugin/deploy/` (ConsolePlugin CR, Deployment, Service, RBAC). The image in the Deployment manifest is set to `CONSOLE_PLUGIN_IMG` via kustomize. The plugin appears in the Administrator perspective under **VCF Migration → Migrations**.

### Undeploy the plugin

```sh
make undeploy-console-plugin
```

### Plugin layout

- `console-plugin/cmd/plugin/` — Go entrypoint; TLS flags, Kube client, server startup
- `console-plugin/pkg/server/` — HTTP server, route registration, static file serving
- `console-plugin/pkg/handlers/` — vSphere API handlers (connect, datacenters, clusters, datastores, networks, resource pools, templates, folders) and SSE events handler
- `console-plugin/web/` — React + TypeScript + PatternFly; `console-extensions.json` for nav/routes; SDK for K8s watch/create
- `console-plugin/deploy/` — ConsolePlugin, Deployment, Service, RBAC, kustomization

## Container Images

Both the operator and console plugin images are built with podman by default. The container tool can be overridden via `CONTAINER_TOOL`.

| Target | Description |
|--------|-------------|
| `make operator-image IMG=...` | Build the operator image |
| `make operator-push IMG=...` | Push the operator image |
| `make console-plugin-image CONSOLE_PLUGIN_IMG=...` | Build the console plugin image |
| `make console-plugin-push CONSOLE_PLUGIN_IMG=...` | Push the console plugin image |

Both `make deploy` and `make deploy-console-plugin` use kustomize to set the image in the respective manifests before applying, so the deployed image always matches the variable you pass.

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/vcf-migration-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/vcf-migration-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing

Contributions are welcome. Please open an issue or PR and follow the code style and conventions described in this repo (see also `AGENTS.md` for build, lint, and test commands).

**NOTE:** Run `make help` for more information on all potential `make` targets.

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
