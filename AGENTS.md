# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Overview

`oci-cloud-controller-manager` is the Oracle Cloud Infrastructure (OCI)
out-of-tree Kubernetes cloud provider. It includes the cloud controller manager
for node and load balancer integration, CSI drivers for OCI storage, flexvolume
components, and volume provisioners.

Most implementation work lives under `cmd/` and `pkg/`. Documentation and
deployment assets live under `docs/`, the root component docs, and `manifests/`.

## Repository Map

- `cmd/`: component entrypoints, including the CCM, CSI drivers, flexvolume
  driver, and volume provisioner binaries.
- `pkg/`: core Go packages for OCI clients, cloud provider behavior, CSI,
  flexvolume, metrics, utilities, and provisioners.
- `manifests/`: Kubernetes YAML and Helm chart assets used to deploy the
  project components.
- `docs/`: development notes and user-facing feature documentation.
- `hack/`: repository scripts for checks, tests, releases, e2e execution, and
  boilerplate validation.
- `test/e2e/cloud-provider-oci/`: OCI/Kubernetes e2e tests and instructions for
  running them against a real cluster.

## Build and Test

- Treat `go.mod` as the source of truth for the required Go version and module
  dependencies.
- Use `make build` to build all configured components into `dist/`.
- Use `make test` for the standard unit test target. It runs the repository test
  script against `cmd` and `pkg`.
- Use `make check` before proposing code changes. It runs gofmt, go vet, and
  golint through the scripts in `hack/`.
- For fast iteration, run focused Go tests such as
  `go test ./pkg/cloudprovider/providers/oci -run TestName -count=1`, then run
  the relevant Make target before finalizing.
- Use `make vendor` only when dependency changes require regenerating
  `vendor/`.
- E2E tests require a real Kubernetes cluster, OCI configuration, `ginkgo`, and
  the environment described in `test/e2e/cloud-provider-oci/README.md`. Do not
  run e2e tests unless the task explicitly requires them and credentials are
  available.

## Coding Guidelines

- Follow existing package boundaries and patterns before introducing new
  abstractions.
- Keep changes scoped to the component being modified. Avoid broad refactors
  unless they are necessary for the requested behavior.
- Run `gofmt` on touched Go files; `make check` verifies formatting, vet, and
  lint expectations.
- Add or update unit tests close to the changed package when behavior changes.
- Do not edit vendored dependencies, generated release directories, or generated
  manifests unless the task is specifically about those artifacts.
- Preserve existing public behavior, annotations, manifest fields, and config
  semantics unless the change intentionally updates them.

## Security and Secrets

- Do not commit OCI credentials, kubeconfigs, private keys, generated
  `cloud-provider.yaml`, `cloud-provider.cfg`, or local config files.
- Check `.gitignore` before adding files that may contain environment-specific
  values.
- Follow `SECURITY.md` for vulnerability handling. Do not open public issues or
  PRs containing undisclosed security details.
- Be careful when changing authentication, authorization, networking, load
  balancer, or storage code; these paths can affect customer infrastructure.

## Contribution Notes

- Follow `CONTRIBUTING.md` for the pull request process.
- Ensure there is an issue for non-trivial fixes or enhancements before opening
  a PR.
- Commits should include the OCA sign-off line, normally by using
  `git commit --signoff`.
- Commit messages should use the `External-ccm:` prefix expected by this
  project.
- Update documentation, samples, and manifests when behavior or user-facing
  configuration changes.
- Keep PR descriptions explicit about what changed and how reviewers can
  validate it.
