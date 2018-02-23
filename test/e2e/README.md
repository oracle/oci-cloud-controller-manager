# E2E Tests

These tests are adapted from the [Service test suite][1] in the Kubernetes core
E2E tests.

## Running

```
$ ginkgo test/e2e -v --progress -- --kubeconfig=$HOME/.kube/config
```


[1]: https://github.com/kubernetes/kubernetes/blob/0cb15453dae92d8be66cf42e6c1b04e21a2d0fb6/test/e2e/network/service.go
