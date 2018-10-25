# E2E Tests

These tests are adapted from the [Service test suite][1] in the Kubernetes core
E2E tests.

## Running

```bash
export SUBNET_OCID=ocid1.subnet.oc1.phx.aaaaaaaanlsnbcixkkchz6n6eznusplxui3xwgb7bsaeucqy4zpehohcb3ra
export MNT_TARGET_OCID=ocid1.mounttarget.oc1.phx.aaaaaa4np2snlxveobuhqllqojxwiotqnb4c2ylefuzaaaaa
export IMAGE=iad.ocir.io/oracle/oci-volume-provisioner
export VERSION="<version under test>"
export OCICONFIG=$(pwd)/oci-config.yaml
export KUBECONFIG="${HOME}/Projects/kubernetes-test-terraform/oci-volume-provisioner-system-test/generated/kubeconfig"
ginkgo \
  -v \
  -progress \
  test/e2e -- \
  --kubeconfig="${KUBECONFIG}" \
  --ociconfig="${OCICONFIG}" \
  --subnet-id="${SUBNET_OCID}" \
  --mnt-target-id="${MNT_TARGET_OCID}" \
  --delete-namespace-on-failure=false \
  --image="${IMAGE}:${VERSION}"
```

## Additional options

| Flag                            | Usage                                                                                                                              | Value  |
|---------------------------------|------------------------------------------------------------------------------------------------------------------------------------|--------|
| `--kubeconfig`                  | Path to Kubeconfig file with authorization and master location information.                                                        | string |
| `--ociconfig`                   | Path to OCIconfig file with cloud provider config.                                                                                 | string |
| `--mnt-target-id`               | Identifies the mount target id for a FSS.                                                                                          | string |
| `--subnet-id`                   | Identifies a subnet to look for a mount target, such that a FSS can be mounted.                                                    | string |
| `--ad`                          | Identifies the availability domain in which the PD resides                                                                         | string |
| `--image`                       | Specifies the container image and version                                                                                          | string |
| `--namespace`                   | Name of an existing Namespace to run tests in.                                                                                     | string |
| `--delete-namespace`            | If true tests will delete namespace after completion. It is only designed to make debugging easier, DO NOT turn it off by default. | bool   |
| `--delete-namespace-on-failure` | If true tests will delete their associated namespace upon completion whether the test has failed.                                  | bool   |
