# CHANGELOG

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
