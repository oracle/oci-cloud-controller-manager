# CHANGELOG

## 0.5

 * Implement `manageRoutes` to enable the cloud route controller
   - This will make sure that `VNIC`s have `SkipSourceDestCheck`
   - Also make sure that all route tables for all instances have the
       appropriate route entries for cluster cidr ranges

## 0.4

 * Implement `loadbalancer.securityListManagementMode: Frontend` which only
   manages security list rules for load balancer ingress. [#180][16]
 * Deprecate `loadbalancer.disableSecurityListManagement` in favour of
   `loadbalancer.securityListManagementMode: None`. [#180][16]
 * Implement `loadbalancer.securityLists` to allow explicit configuration of the
   security lists that the CCM manages on a per-subnet basis [#164][17].
 * Implement support for [instance principles][19] authentication [#155][18]

## 0.3.2

 * [BUG] Fix panic when EnsureLoadBalancer() called with 0 Nodes [#176][11]
 * [BUG] Fix panic when GetInstanceByNodeName() encountered an instance without
   either a public IP or a hostname [#167][14]
 * [BUG] Fix regression where compartment OCID was no longer looked up from
   metadata when not provided in cloud-provider config [#168][15]
 * Deprecate cloud-provider config property `auth.key_passphrase` replacing it
   with `auth.passphrase` [#142][12]
 * Deprecate cloud-provider config property `auth.compartment` replacing it
   with `compartment` [#170][13]

## 0.3.1

 * Remove redundant `--cluster-cidr` flag from DaemonSet [#163][10]

## 0.3.0

 * Create load balancers using a single OCI API request significantly reducing fresh load balancer creation time [[#148][6]]
 * Fix failure to delete security list rules when deleting a Service type=LoadBalancer or changing its NodePort(s) [[#151][7]]
 * Move to ConfigMap based leader election [[#160][8]]
 * Fix potential panic when a load balancer had no backends [[#157][9]]

## 0.2.1

 * Add OCI proxy support [[#135][5]]
 * Add security rules for a services health check port [[#125][4]]

## 0.2.0

 * Wercker release process [[#131][3]]

## 0.1.2

 * Fixes a bug where existing security list rules with no destination port range
   set would cause the CCM to fail when update security lists whilst ensuring
   load balancer state [[#112][2]]

## 0.1.1

 * Fix failure to update load balancer listener port when frontend service port changed [[#92][1]]

## 0.1.0

 * Implement `cloudprovider.Instances` interface - keeps track of node state within an OCI Kubernetes cluster.
 * Implement `cloudprovider.LoadBalancer` interface - enables create, update, and delete of OCI load balancers for services of type `LoadBalancer`.
 * Implement `cloudprovider.Zones` interface - provides OCI region information for cluster nodes.

[1]: https://github.com/oracle/oci-cloud-controller-manager/issues/92
[2]: https://github.com/oracle/oci-cloud-controller-manager/issues/112
[3]: https://github.com/oracle/oci-cloud-controller-manager/issues/131
[4]: https://github.com/oracle/oci-cloud-controller-manager/issues/125
[5]: https://github.com/oracle/oci-cloud-controller-manager/issues/135
[6]: https://github.com/oracle/oci-cloud-controller-manager/issues/148
[7]: https://github.com/oracle/oci-cloud-controller-manager/issues/151
[8]: https://github.com/oracle/oci-cloud-controller-manager/issues/160
[9]: https://github.com/oracle/oci-cloud-controller-manager/issues/157
[10]: https://github.com/oracle/oci-cloud-controller-manager/issues/163
[11]: https://github.com/oracle/oci-cloud-controller-manager/issues/176
[12]: https://github.com/oracle/oci-cloud-controller-manager/issues/142
[13]: https://github.com/oracle/oci-cloud-controller-manager/issues/170
[14]: https://github.com/oracle/oci-cloud-controller-manager/issues/167
[15]: https://github.com/oracle/oci-cloud-controller-manager/issues/168
[16]: https://github.com/oracle/oci-cloud-controller-manager/issues/180
[17]: https://github.com/oracle/oci-cloud-controller-manager/issues/164
[18]: https://github.com/oracle/oci-cloud-controller-manager/issues/155
[19]: https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Tasks/callingservicesfrominstances.htm
