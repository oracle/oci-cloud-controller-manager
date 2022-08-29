# OCI Cloud Controller Manager (CCM)

`oci-cloud-controller-manager` is a Kubernetes Cloud Controller Manager
implementation (or out-of-tree cloud-provider) for [Oracle Cloud
Infrastucture][1] (OCI).

[![wercker status](https://app.wercker.com/status/17a52304e0309d138ad41f7ae9f9ea49/s/master "wercker status")](https://app.wercker.com/project/byKey/17a52304e0309d138ad41f7ae9f9ea49)
[![Go Report Card](https://goreportcard.com/badge/github.com/oracle/oci-cloud-controller-manager)](https://goreportcard.com/report/github.com/oracle/oci-cloud-controller-manager)

## Introduction

External cloud providers were introduced as an _Alpha_ feature in Kubernetes
1.6 with the addition of the [Cloud Controller Manager][2] binary. External
cloud providers are Kubernetes (master) controllers that implement the
cloud-provider specific control loops required for Kubernetes to function.

This functionality is implemented in-tree in the `kube-controller-manger` binary
for _existing_ cloud-providers (e.g. AWS, GCE, etc.), however, in-tree
cloud-providers have entered maintenance mode and _no additional providers will
be accepted_. Furthermore, there is an ongoing effort to remove all existing
cloud-provider specific code out of the Kubernetes codebase.

## Compatibility matrix

|           | Min Kubernetes Version | Max Kubernetes Version |
|-----------|------------------------|------------------------|
| \>=v 0.11 | v1.16                  | v1.18                  |
| \>=v 0.12 | v1.18                  | v1.21                  |
| \>=v 0.13 | v1.19                  | v1.21                  |
| v1.19.12  | v1.19                  | v1.21                  |
| v1.22.0   | v1.22                  | -                      |
| v1.23.0   | v1.23                  | -                      |
| v1.24.0   | v1.24                  | -                      |



Note: 
Versions older than v0.13.0 are no longer supported, new features / bug fixes will be available in v0.13.0 and later. 

## Implementation
 Currently `oci-cloud-controller-manager` implements:

 - NodeController - updates nodes with cloud provider specific labels and
   addresses, also deletes kubernetes nodes when deleted from the
   cloud-provider.
 - ServiceController - responsible for creating load balancers when a service
   of `type: LoadBalancer` is created in Kubernetes.

 Additionally, this project implements a container-storage-interface, a flexvolume driver and a flexvolume provisioner for Kubernetes clusters running on Oracle Cloud Infrastructure (OCI).

## Setup and Installation

To get the CCM running in your Kubernetes cluster you will need to do the
following:

 1. Prepare your Kubernetes cluster for running an external cloud provider.
 2. Create a Kubernetes secret containing the configuration for the CCM.
 3. Deploy the CCM as a [DaemonSet][4].

Note: For the setup and installation of [flexvolume driver](flex-volume-driver.md), [flexvolume provisioner](flex-volume-provisioner.md) and [container-storage-interface](container-storage-interface.md) please refer linked resources.

### Preparing Your Cluster

To deploy the Cloud Controller Manager (CCM) your cluster must be configured to
use an external cloud-provider.

This involves:
 - Setting the `--cloud-provider=external` flag on the `kubelet` on **all
   nodes** in your cluster.
 - Setting the `--provider-id=<instanceID>` flag on the `kubelet` on **all
   nodes** in your cluster. 
   Where `<instanceID>` is the [instance ocid][11] of a node (unique for each node).
 - Setting the `--cloud-provider=external` flag on the `kube-controller-manager`
   in your Kubernetes control plane.

**Depending on how kube-proxy is run you _may_ need the following:**

- Ensuring that `kube-proxy` tolerates the uninitialised cloud taint. The
  following should appear in the `kube-proxy` pod yaml:

```yaml
- effect: NoSchedule
  key: node.cloudprovider.kubernetes.io/uninitialized
  value: "true"
```

If your cluster was created using `kubeadm` >= v1.7.2 this toleration will
already be applied. See [kubernetes/kubernetes#49017][5] for details.

Remember to restart any components that you have reconfigured before continuing.

### Authentication and Configuration

An example configuration file can be found [here][7]. Download this file and
populate it with values specific to your chosen OCI identity and tenancy.
Then create the Kubernetes secret with the following command:

For CCM -
```bash
$ kubectl  create secret generic oci-cloud-controller-manager \
     -n kube-system                                           \
     --from-file=cloud-provider.yaml=provider-config-example.yaml
```
Note that you must ensure the secret contains the key `cloud-provider.yaml`
rather than the name of the file on disk.

### Deployment

Deploy the controller manager and associated RBAC rules if your cluster
is configured to use RBAC (replace ? with the version you want to install to):

```bash
$ export RELEASE=?
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-cloud-controller-manager-rbac.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-cloud-controller-manager.yaml
```

Check the CCM logs to ensure it's running correctly:

```bash
$ kubectl -n kube-system get po | grep oci
oci-cloud-controller-manager-ds-k2txq   1/1       Running   0          19s

$ kubectl -n kube-system logs oci-cloud-controller-manager-ds-k2txq
I0905 13:44:51.785964       7 flags.go:52] FLAG: --address="0.0.0.0"
I0905 13:44:51.786063       7 flags.go:52] FLAG: --allocate-node-cidrs="false"
I0905 13:44:51.786074       7 flags.go:52] FLAG: --alsologtostderr="false"
I0905 13:44:51.786078       7 flags.go:52] FLAG: --cloud-config="/etc/oci/cloud-config.cfg"
I0905 13:44:51.786083       7 flags.go:52] FLAG: --cloud-provider="oci"
```

## Upgrade

The following example shows how to upgrade the CCM, FVP, FVD and CSI from an older version (replace ? with the version you're upgrading to):

```bash
$ export RELEASE=?
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-cloud-controller-manager-rbac.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-cloud-controller-manager.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-volume-provisioner.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-volume-provisioner-rbac.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-flexvolume-driver.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-flexvolume-driver-rbac.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-csi-controller-driver.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-csi-node-driver.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-csi-node-rbac.yaml
```

## Examples

 - [Service `type: LoadBalancer` basic NGINX example][8]
 - [Service `type: LoadBalancer` NGINX SSL example][9]

## Development

See [DEVELOPMENT.md](docs/development.md).

## Support

If you think you've found a bug, please [raise an issue][3].

## Contributing

`oci-cloud-controller-manager` is an open source project. See [CONTRIBUTING](CONTRIBUTING.md) for
details.

Oracle gratefully acknowledges the contributions to this project that have been made
by the community.

## Upcoming Releases

| Release               | Expected Release Date |
|-----------------------|-----------------------|
| Support for K8s v1.24 | August 2022           |

## License

Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

`oci-cloud-controller-manager` is licensed under the Apache License 2.0.

See [LICENSE](LICENSE) for more details.

[1]: https://cloud.oracle.com/iaas
[2]: https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/
[3]: https://github.com/oracle/oci-cloud-controller-manager/issues/new
[4]: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
[5]: https://github.com/kubernetes/kubernetes/pull/49017
[6]: https://kubernetes.io/docs/concepts/containers/images/#creating-a-secret-with-a-docker-config
[7]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/manifests/provider-config-example.yaml
[8]: https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/tutorial.md
[9]: https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/tutorial-ssl.md
[10]: https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/rate-limiter-configuration.md
[11]: https://docs.cloud.oracle.com/en-us/iaas/Content/Compute/Concepts/computeoverview.htm#two
