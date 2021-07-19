# E2E Tests

These tests are adapted from the [Service test suite][1] in the Kubernetes core
E2E tests.

## Pre-requisite
You need to have [ginkgo][2] installed and configured.

## Running the tests locally

If you already have a cluster and wish to run the tests on your existing cluster.

Define the following environment variables in your shell environment - 

```bash
# ADLOCATION example is IqDk:US-ASHBURN-AD-1
export ADLOCATION=<adlocation>
export CLUSTER_KUBECONFIG=<file path to your cluster's kubeconfig>
export CLOUD_CONFIG=<path that points to cloud-provider.yaml for your cluster>
export CMEK_KMS_KEY=<CMEK ocid>
```

Then run

```bash
make run-ccm-e2e-tests-local
```

An example of cloud-provider.yaml - See [provider-config-example.yaml](../../../manifests/provider-config-example.yaml)
For accessing you cluster's kubeconfig refer [organize-cluster-access-kubeconfig][3]

NOTE: Test suite will fail if executed behind a `$HTTP_PROXY` that returns a
200 OK response upon failure to connect.

## Additional option to specify test image pull repo

The tests use below images - 
*   nginx:stable-alpine
*   netexec:1.1
*   centos:latest
*   busybox:latest

By default, public images are used. But if your Cluster's environment cannot access above public images then below option can be used to specify an accessible repo.

```bash
IMAGE_PULL_REPO="accessiblerepo.com/repo/path/" make run-ccm-e2e-tests-local
```

Note: Above listed <IMAGE>:<TAG> should be available in the provider repo path and should be accessible.

## Additional Debug Options when running tests on existing cluster

Additional seclist count based sanity checks can be applied during e2e testing
by providing the appropriate seclist ocids. Both must be supplied.
If you wish to use the tests in debug mode and want to look at the namespaces created by tests you can provide the option to keep the namespaces 
```bash
--delete-namespaces=false
```
By default the namespaces will be deleted. So if you use above option, then you should delete the resources manually after investigation is complete.
These values can be specified as command line parameters.

```bash
$ ginkgo -v -progress test/e2e/cloud-provider-oci -- \
    --clusterkubeconfig=$CLUSTER_KUBECONFIG \
    --delete-namespace=false \
    --cloud-config=$CLOUD_CONFIG \
    --adlocation=$ADLOCATION \
    --ccm-seclist-id=ocid1.securitylist.$ccmloadblancerid \
    --k8s-seclist-id=ocid1.securitylist.$k8sworkerid
    --cmek-kms-key=${CMEK_KMS_KEY} 
```

---

### Running subsets of the cloud-provider-oci-e2e-tests (focused tests)

You can run a subset of the tests manually by setting the 'FOCUS' environment variable to be the regular expressions matching the 'Ginkgo' descriptions of the test you want to run.

Please see 'test/e2e/cloud-provider-oci' to see what description tags are available.
The broad category of tags is "ccm" and "storage" look at [list of tests](ListOfTests.md)

```bash
export FOCUS=\[cloudprovider\]
```
They can be passed directly to the make target as well:

```bash
FOCUS="\[cloudprovider\]" make run-ccm-e2e-tests-local
```

Another way you can run a subset of the tests manually is by setting the 'FOCUS' environment variable
to be the regular expression for the test files you want to run. In this case you must also set an
environment variable 'FILES=true' so Gingko knows to interperate the expression in that way.

```bash
export FOCUS="instances.go"
export FILES="true"
```

They can be passed directly to the make target as well:

```bash
make run-ccm-e2e-tests-local FOCUS="load_*" FILES="true"
```

---

[1]: https://github.com/kubernetes/kubernetes/blob/0cb15453dae92d8be66cf42e6c1b04e21a2d0fb6/test/e2e/network/service.go
[2]: https://onsi.github.io/ginkgo/
[3]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
