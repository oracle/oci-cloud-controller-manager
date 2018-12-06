## cloud-provider-oci Configuration Options

This document outlines all of the configuration options available for the Oracle
Cloud Infrastructure (OCI) Kubernetes cloud-provider (cloud-provider-oci).

## Contents

- [cloud-provider-oci Configuration Options](#cloud-provider-oci-configuration-options)
- [Contents](#contents)
- [Cloud Configuration File](#cloud-configuration-file)
    - [Sample configuration](#sample-configuration)
    - [Configuration Options](#configuration-options)
        - [Top Level](#top-level)
            - [Top Level Required Parameters](#top-level-required-parameters)
            - [Top Level Optional Parameters](#top-level-optional-parameters)
        - [Auth](#auth)
            - [Auth Required Parameters](#auth-required-parameters)
            - [Auth Optional Parameters](#auth-optional-parameters)
        - [Loadbalancer](#loadbalancer)
            - [Loadbalancer Optional Parameters](#loadbalancer-optional-parameters)


## Cloud Configuration File

Kubernetes knows how to interact with OCI via the file `cloud-provider.yaml`.
It is a standard YAML file that provides Kubernetes with user authentication
credentials and additional configuration specific to your OCI account.

### Sample configuration

This is an example of a typical configuration that touches the values that most
often need to be set. It provides details for how to authenticate with OCI via a
standard IAM user and configures load balancer provisioning.

```yaml
auth:
  region: us-ashburn-1
  tenancy: ocid1.tenancy.oc1..
  user: ocid1.user.oc1..
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    <snip>
    -----END RSA PRIVATE KEY-----
  fingerprint: 8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74

compartment: ocid1.compartment.oc1..
vcn: ocid1.vcn.oc1..

loadBalancer:
  subnet1: ocid1.subnet.oc1.iad..
  subnet2: ocid1.subnet.oc1.iad..
```

### Configuration Options

The OCI Cloud Provider offers a wide range of configuration options for each
service it supports. The currently available configuration sections include:

 - [Top Level](#top-level)
 - [Auth](#auth)
 - [Load Balancer](#load-balancer)

#### Top Level

These configuration options for the OCI cloud provider relate to its global
configuration and should appear at the top level of your `cloud-provider.yaml`
file.

##### Top Level Required Parameters

- `compartment`: The OCID of the OCI Compartment within which your Kubernetes
  cluster resides.
- `vcn`: The OCID of the Virtual Cloud Network (VCN) within which your
  Kubernetes cluster resides.

##### Top Level Optional Parameters

- `useInstancePrincipals`: `true`/`false`. If enabled use [Instance Principals][1]
  authentication for communication with OCI services.

#### Auth

These configuration options for the OCI cloud provider relate to authentication
with OCI services via an IAM user and OCI API key. For more information please
see [here][2].

**NOTE**: If `useInstancePrincipals` is set to `true` then all auth fields are
optional.

##### Auth Required Parameters

 - `region`: The identifier for the OCI Region in which your Kubernetes cluster
   resides (e.g. `us-ashburn-1`).
 - `tenancy`: The OCID of your OCI Tenancy.
 - `user`: The OCID of the IAM user as which cloud-provider OCI will
   authenticate with OCI services.
 - `key`: The PEM encoded OCI API private key associated with the `user` as
   which cloud-provider OCI will authenticate with OCI services.
 - `fingerprint`: The fingerprint of the OCI API public key.

##### Auth Optional Parameters

 - `regionKey`: The short identifier for the OCI Region in which your Kubernetes
   cluster resides (e.g. `iad`).
 - `passphrase`: The passphrase if the provided private key is encrypted.

#### Load balancer

These configuration options for the OCI cloud provider relate to provisioning
loadbalancers via the cloud controller manager.

##### Loadbalancer Optional Parameters

 - `disabled`:  `true`/`false`. Disables the creation of load balancers.
 - `securityListManagementMode`: `All` (default), `Frontend`, and `None`.
   Defines how the oci-cloud-controller-manager manages Security Lists to enable
   traffic to ingress/egress the Kubernetes cluster via loadbalancers.

   Management mode `All` enables management of all required Security List rules
   for traffic to ingress/egress the Kubernetes cluster via loadbalancers.

   Management mode `Frontend` enables management of Security List rules for
   loadbalancer ingress only.

   Management mode `None` disables Security List management by the
   oci-cloud-controller-manager completely. This requires that the user
   has setup a rule that allows inbound traffic to the appropriate ports
   for the kube proxy health check port, node port ranges, and health check port
   ranges. E.g. 10.82.0.0/16 30000-32000.
 - `subnet1`: The OCID of the first of two required Subnets to which
   loadbalancers will be attached.
 - `subnet2`: The OCID of the second of two required Subnets to which
   loadbalancers will be attached.
 - `securityLists`: A map of Subnet OCIDs to Security List OCIDs to enable
   configuration of which Security Lists the oci-cloud-controller-manager
   manages when creating ingress/egress rules to enable traffic to
   ingress/egress the Kubernetes cluster via loadbalancers.


[1]: https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm
[2]: https://docs.cloud.oracle.com/iaas/Content/API/Concepts/apisigningkey.htm
