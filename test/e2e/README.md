# E2E Tests

These tests are adapted from the [Service test suite][1] in the Kubernetes core
E2E tests.

## Running

```bash
$ ginkgo -v -progress test/e2e -- --kubeconfig=${HOME}/.kube/config --delete-namespace=false
```

NOTE: Test suite will fail if executed behind a `$HTTP_PROXY` that returns a
200 OK response upon failure to connect.

## Additional Options

Additional seclist count based sanity checks can be applied during e2e testing 
by providing the appropriate seclist ocids. Both must be supplied.

```bash
export CCM_SECLIST_ID="ocid1.securitylist.$ccmloadblancerid"
export K8S_SECLIST_ID="ocid1.securitylist.$k8sworkerid"
```

Alternatively, these values can be specified as command line parameters.

```bash
$ ginkgo -v -progress test/e2e -- \
    --kubeconfig=${HOME}/.kube/config \
    --delete-namespace=false \
    --ccm-seclist-id=ocid1.securitylist.$ccmloadblancerid \
    --k8s-seclist-id=ocid1.securitylist.$k8sworkerid
```


[1]: https://github.com/kubernetes/kubernetes/blob/0cb15453dae92d8be66cf42e6c1b04e21a2d0fb6/test/e2e/network/service.go
