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
        MIIEowIBAAKCAQEA4KpLGy/BLbph55HMjWLxCO657DLQTk4o+WWPi1+5oeAUVgyh
        kdvPR22jn9HiAL9jKv7PR3/OdHSp/6E3d05htksI7Tct4M/eWVMGRIzoMJvpJ99e
        ZP7MtQT9yknbJDSJoibSwLmPoInnPE/WbcgrTKSAfNURK0bKw1tnLd85qt7zdLI3
        g6O/14Bsmf+ovGiQHP6oiTuC4l3D8eTLlKdSrRVqZXhdvslpZU8MtNB8pPHMB4GZ
        R6HccBi7TJY7kkNg+5flRBTdYL8bvaji3zxSlvawvet+bJmEtApkUoLnovLCviVp
        NVTJZb5iQxMJLZlDJJT/ruq+HMJ3PiiYFOjFVwIDAQABAoIBAQDNkiT9MFoj/Hpf
        SOKRsKn60W3gObKvJAeMBKkvD50tCHuzLQWeEDJ/GkxxDbwtkPItwlBqDQEdQC7Z
        UGwPR/JSuh/l5uqc3beHpleC3CgNamwSZunZoegv7uxGcAQMAeK6M6n+XQyWCflD
        D46Wj2VHUPKcxt1Z6wHXdchYifwbYwUNA3hOlRJK3ODgk/X6UjTGb3+gpY3qU4kX
        Iz5L1ekCSgVIPBFVwdZQUyUC7+iIySaK+qcmEEx/UwOZ6uxhcmRzca31cjeaRS4H
        pUjrl/aqLIW57E2MQ/vSzfQn7kEGBOrS0RjHZgq9u4Qdq6EkjHj3fenKpwWB7S1z
        4t0PpinJAoGBAPRmxAcCd88EhWh5HhN+RWjmXdDCOmZ0yXbxxVBTQtK5pPnP8I9A
        3Jd2ughHk7dFBvgKbHkVsyWgAk8zRZdD2hkQBOXvoeJF2scmvgFUBs1otf6xiFsf
        IC0I8A/wXn3IHmyrG7xmPAtHWKvTTAFg7IjIIofcX7cuzMeLXEUMvLQdAoGBAOtT
        wJCtPTNs4c3vhO4gba98c30U3tHmbLVKJXGEeZkSv3/ez5eIiYBJTzwLB2+ppy8j
        2lYsdkLvsoyKF3LUwyt0gsX+AU9DJ2dmSJZ3E67UHsY6+qog5QlYfWWD8mKWeE9L
        2r0rhG6l0WHR15LdvVc9MJ8e3YVUvNJJJJhQ2v0DAoGAAosXOyNxb7wST1YDVBya
        SE8tZsC+rtZESnKVpRJYvayk5NyfGj6IjSL1KKTmCqAzRF2HZ3MsXBXgMEbOUJaq
        LFyYUHQ/8QTdE/l5PLZNI9IVIsNiMeCPCyjuppvPv+tXNbZKIZnGwi9J4u/d+J2z
        mHDMuzE15cgc5W6z1Rwe0pkCgYBzRwvF05dvYZ8bqoGLxQb2OBi65UZhvGb0R+Yf
        va1zduOoWBWJPbFdzoup9h0mbg0f4ohKPm2QTKtCfUMPVXpmByUoqE0r7tGWrVxR
        mPNjaTXKFYpFXOfVtCt5VzGdaeh1r8rvcCnnqgLv0EOyBj2CRs9So2QQtHnq6Tms
        A6/C0QKBgAw8IsCnkNoZujCEOR/6ZHbK3eeyAs2yuJumsjYYosIGZ/bzsXTpfzAw
        bs45GZxrW67zB/0HA7bVWS9ZkCVflHI2uBCFofm+y55IAzg9/c1xYU19PA3KRxHZ
        D/yEDdXVK/lIzNt7kIMFhtoYGrwv1JQGfK5Wh2bi+AwbBDZ45/17
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

The e2e test suite requires the version of the CCM under test to be installed in
the cluster referenced via the `--kubeconfig` flag.

```console
$ ginkgo -v -progress test/e2e  -- --kubeconfig=$HOME/.kube/config --cloud-config="$(pwd)/cloud-provider.yaml"
```

[1]: https://www.docker.com/
[2]: https://github.com/golang/dep
[3]: https://golang.org/
[4]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
