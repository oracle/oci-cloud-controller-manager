# Kubernetes Cloud Controller Manager (CCM)

This project is an exernal Kubernetes cloud provider for [Oracle Cloud
Infrastucture][1].

[![wercker status](https://app.wercker.com/status/17a52304e0309d138ad41f7ae9f9ea49/s/master "wercker status")](https://app.wercker.com/project/byKey/17a52304e0309d138ad41f7ae9f9ea49)

## Introduction

External cloud providers were introduced as an _Alpha_ feature in Kubernetes
1.6. External cloud providers are Kubernetes (master) controllers that implement
the cloud-provider specific control loops required for Kubernetes to function.

This functionality is implemented in-tree in the `kube-controller-manger` binary
for _existing_ cloud-providers (e.g. AWS, GCE, etc.), however, in-tree
cloud-providers have entered maintenance mode and _no additional providers will
be accepted_.

## Compatibility matrix

|                                  | Kubernetes < 1.7.6 | Kubernetes 1.7.6 |
|----------------------------------|--------------------|------------------|
| oci-cloud-controller-manager 0.1 | †                  | ✓                |

Key:

* `✓` oci-cloud-controller-manager is fully compatible.
* `†` oci-cloud-controller-manager is compatible but requires the
      `--provider-id` flag to be set on the kubelet of all nodes in the
      cluster.

## Cloud Controller Manager

The cloud-controller-manager allows cloud vendors code and the Kubernetes core
to evolve independent of each other. In prior releases, the core Kubernetes code
was dependent upon cloud-provider-specific code for functionality. In future
releases, code specific to cloud vendors should be maintained by the cloud
vendor themselves, and linked to cloud-controller-manager while running
Kubernetes.

The following controllers have cloud provider dependencies:

- **Node Controller**: For checking the cloud provider to determine if a node
  has been deleted in the cloud after it stops responding.
- **Route Controller**: For setting up routes in the underlying cloud
  infrastructure.
- **Service Controller**: For creating, updating and deleting cloud provider
  load balancers.
- **Volume Controller**: For creating, attaching, and mounting volumes, and
  interacting with the cloud provider to orchestrate volumes.

## Setup and Installation

See [INSTALL.md](docs/install.md).

## Development

See [DEVELOPMENT.md](docs/development.md).

## Support

If you need support, start with the documentation. If you still have problems
[raise an issue][2] or contact odx_kubernetes_gb_grp@oracle.com.

[1]: https://cloud.oracle.com/iaas
[2]: https://github.com/oracle/oci-cloud-controller-manager/issues/new
