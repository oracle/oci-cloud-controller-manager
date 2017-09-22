# Development documentation

The CCM has a build system based on `make` and [Docker][1]. Dependencies are
managed using [Glide][2].

## Running locally
 1. Ensure you have the aforementioned development tools installed as well as
    as well as the latest release of [Go][3].

 2. Create a `cloud-provider.cfg` file in the root of the repository containing
    the appropriate configuration (making sure to reference your OCI API signing
    key path). e.g.

    ```ini
    [Global]
    user = ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q
    fingerprint = 8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74
    key-file = /Users/apryde/.oci/oci_api_key.pem
    tenancy = ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq
    compartment = ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq
    [LoadBalancer]
    subnet1 = ocid1.subnet.oc1.phx.aaaaaaaasa53hlkzk6nzksqfccegk2qnkxmphkblst3riclzs4rhwg7rg57q
    subnet2 = ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq
    ```
 3. Ensure you have [`$KUBECONFIG`][4] to the Kubernetes configuration file for
    your cluster.

 4. Execute `make run-dev`

[1]: https://www.docker.com/
[2]: https://glide.sh/
[3]: https://golang.org/
[4]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
