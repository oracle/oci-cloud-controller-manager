# E2E Tests

These tests are adapted from the [Service test suite][1] in the Kubernetes core
E2E tests.

## Running


```bash
$ ginkgo -v -progress test/e2e -- --kubeconfig=${HOME}/.kube/config --delete-namespace=false
```

NOTE: Test suite will fail if executed behind a `$HTTP_PROXY` that returns a
200 OK response upon failure to connect.


[1]: https://github.com/kubernetes/kubernetes/blob/0cb15453dae92d8be66cf42e6c1b04e21a2d0fb6/test/e2e/network/service.go
