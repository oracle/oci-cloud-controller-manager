# Development documentation

The CCM has a simple build system based on `make`. Dependencies are managed
using [`dep`][2].

## Running locally
 1. Ensure you have the aforementioned development tools installed as well as
    as well as the latest release of [Go][3].

 2. Create a `cloud-provider.yaml` file in the root of the repository containing
    the appropriate configuration. e.g.

    ```yaml
    auth:
      region: us-phoenix-1
      tenancy: ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq
      user: ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q
      key: |
        -----BEGIN RSA PRIVATE KEY-----
        <snip/>
        -----END RSA PRIVATE KEY-----
      fingerprint: 97:84:f7:26:a3:7b:74:d0:bd:4e:08:a7:79:c9:d0:1d
    compartment: ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq
    loadBalancer:
      disableSecurityListManagement: false
      subnet1: ocid1.subnet.oc1.phx.aaaaaaaasa53hlkzk6nzksqfccegk2qnkxmphkblst3riclzs4rhwg7rg57q
      subnet2: ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq
    ```
 3. Ensure you have [`$KUBECONFIG`][4] to the Kubernetes configuration file for
    your cluster.

 4. Execute `GOOS=darwin make run-dev`

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

See [here][5].

[1]: https://www.docker.com/
[2]: https://github.com/golang/dep
[3]: https://golang.org/
[4]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
[5]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/test/e2e/cloud-controller-manager/README.md
