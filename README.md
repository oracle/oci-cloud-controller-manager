# Kubernetes Cloud Controller Manager (CCM)

This project is an exernal Kubernetes cloud provider for Oracle's Bare Metal
Cloud (BMCS).

[![build status](https://gitlab-odx.oracle.com/odx/bmc-cloud-controller-manager/badges/master/build.svg)](https://gitlab-odx.oracle.com/odx/bmc-cloud-controller-manager/commits/master)

## Introduction

External cloud providers were introduced as an _Alpha_ feature in Kubernetes
1.6. External cloud providers are Kubernetes (master) controllers that implement
the cloud-provider specific control loops required for Kubernetes to function.

This functionality is implemented in-tree in the `kube-controller-manger` binary
for _existing_ cloud-providers (e.g. AWS, GCE, etc.), however, in-tree
cloud-providers have entered maintenance mode and _no additional providers will
be accepted_.

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
[raise an issue][1] or contact odx_kubernetes_gb_grp@oracle.com.

[1]: https://gitlab-odx.oracle.com/odx/bmc-cloud-controller-manager/issues/new?issue%5Bassignee_id%5D=&issue%5Bmilestone_id%5D=
