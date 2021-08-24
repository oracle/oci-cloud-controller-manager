# Development documentation

The CCM has a simple build system based on `make`. Dependencies are managed
using [`go mod`][2].

## Setup
 1. Ensure you have the aforementioned development tools installed as well as the release 1.15.x of [Go][3].

 2. Clone the repo under `/src/github.com/oracle/`
 
 3. make vendor
 
## Build locally
 1. `make build-local $COMPONENT`
 2. `make image $COMPONENT `
 Note: `$COMPONENT` is optional. If you don't specify all components image will be built and placed under dist folder.

## DaemonSet manifests

You can template `manifests/cloud-controller-manager/oci-cloud-controller-manager.yaml`
using `make manifests`. This enables running specific versions of the CCM with
the proviso that the version has been pushed to Github, the CI pipeline has
passed, and `HEAD` is pointed to the commit in question. You can then execute
the following to run the CCM as a DaemonSet (RBAC optional):

```console
$ kubectl apply -f dist/oci-cloud-controller-manager.yaml
$ kubectl apply -f dist/oci-cloud-controller-manager-rbac.yaml
```

## Running the e2e tests
See [README.md](../test/e2e/cloud-controller-manager/README.md)

[1]: https://www.docker.com/
[2]: https://github.com/golang/go/wiki/Modules
[3]: https://golang.org/

