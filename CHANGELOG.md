# CHANGELOG

## v0.13.0
### Major Update
k8s vendor dependencies upgraded to 1.19 from 1.16. Cluster with kubernetes version 1.18 or lesser will not be compatible with any of the components in this release.

### Min required K8s version
1.19

### Features
* ARM support with multiarch images
* OKE (CSI) Expand Volume support. https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/expand-block-volume-using-csi.md
* Support for tagging OCI LB and BV
* Support Static Provisioning of FSS using CSI Driver with In-transit Encryption. https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/statically-provisioned-fss-using-csi.md

### Enhancements
* Add arm image support and minor bug fix
* fixed encryption disabled fss tests
* Replaced lbOCID with loadBalancerID in log objects
* Removed hardcoded function and race condition
* Addition of Backendset size dimension to LB metrics
* Added health checks of nodes from pods using a job
* added test to check data integrity on pod restart

### Bug Fixes
* CSI Block Volume - Unexpected formatting when staging target path does not exist and mount fails
* Bug fix for block volume encryption
* Adding error condition when in transit is on and attachment type is iscsi
* Clean exit if mount point is not found and mitigating remove directory error

## v0.13.0-alpha1

## 0.12.0
#### Important points
* NSG Support
* Reserved IP Support
* Kubernetes 1.19 and 1.20 Support
* Bug fixes
#### Summary of changes
* Fix for pod deletion and recreation in fvd test
* removed TCP health check for SSL enabled traffic
* CCM should use LB bulk update API to add /delete backend in backendset
* CCM should stop fetching node and lb subnet in time of LB deletion if CCM is not managing the security list.
* Make all RPCs mutually exclusive
* CCM support for public reserved IP
* Introduce paging to ListVolumes, no limit, filter by display-name
* [CSI] Create targetPath if not already created
* Support for Associating NetworkSecurityGroups to LoadBalancer
* Add resourceName to provision failures to dedupe metrics
* Enable metrics for update operations
* Enable IMDS server lookup

## 0.11.0

* Handle Volume Attachments created in the non-cluster compartment
* Fix nil pointer dereference when node not found by IP
* Add support for updating LB Shape and provisioning of Flexible shape LBs

## 0.10.0

* Service annotation oci-load-balancer-internal: false is not honoured
* CSI - Wait for volume to reach DETACHED state before attempting to attach
* Add unit tests for CSI
* Fix Backup Restore E2E test for multi Attach Error
* Drop iSCSI package installation and add util-linux to the cloud-provider-oci image

## 0.9.0
##### Major changes -

* CSI support
* Regional Subnet support via LB annotation
* Health check config support for backend sets
* Support LB Service creation if nodepool belongs to a different compartment than cluster's compartment
* Bug fixes

##### Detailed changes -

* Kubernetes 1.18 compatibility
* Fix node info controller to trim providerID prefix to support backward compatibility to non-oke clusters.
* CCM should support LB Service creation if nodepool belongs to a different compartment
* CCM should set node address properly to cross compartment node and Fault Domain Label to all the Nodes.
* Opensource CSI in OGHO
* remove wercker references
* CCM is not honouring oci-load-balancer-connection-idle-timeout annotation during update event
* Update seclist when lb service nodeport changes
* Ensure LB SecList is synced if there are no Backend/Backendset/Listerner changes
* merge internal repo commits to oss
* oci-cloud-provider should retry on all retryable errors
* Regional Subnet via LB annotation re-work. Remove subnets annotation and keep old annotaion.
* Support health check config for backend sets.
* Customer should be able to edit the service to add/update listener even if the customer uses the service.beta.kubernetes.io/oci-load- balancer-backend-protocol: "HTTP" 2. When customers add service.beta.kubernetes.io/oci-load-balancer-backend-protocol: "HTTP" in existing service, CCM should update the existing listener with the new protocol( i.e only change the protocol of the listener, not create and delete the listerner)

## 0.8.0
* Kubernetes version 1.14, 1.15, 1.16 and 1.17 support.
* Change documentation as per new release. fix rbac manifest and readme compatibility matrix
* add/delete/update backend servers instead of updating the entire backendsets
* Fix the documentation around how to run e2e tests for CCM
* Fix issues with ccm e2e tests.
* CCM is not honouring oci-load-balancer-connection-idle-timeout annotation during update event
* Change log level to debug for config file not found on worker node
* Fixing CCM build target
* Resolve dependency conflict for Sirupsen by bumping up the version of docker/docker
* CCM should ignore nodes for backends that aren't setup yet.
* move to go mod
* Avoid CCM panic due to index out of range
* update oci-go-sdk
* correction to boilerplate header validation and remove e2e test automation
* Enable provisioning BV from another BV
* Deploy single container with CCM and FVP
* Cert rotation should avoid listener recreation
* Fix test to avoid appending secret name to listeners name

## 0.7.0

* Support assigning SSL certs to BackendSets [#243][30]

## 0.6.1

* Support load balancer listeners with http protocol via annotation [#239][28]
* Prevent leaking security list rules when updating a service [#238][29]

## 0.6.0

* Allow dynamic configuration of security list management via annotations. [#217][23]
* Migrate to Zap logger. [#223][24]
* Runtime checks around providerID. [#228][25]
* Support request rate limiting. [#201][26]
* Bump dependencies to Kubernetes 1.11.1. [#234][27]

## 0.5.1

* Fix panic on startup when security lists aren't managed. [#213][22]
* Update oci-go-sdk to 2.0.0 which includes support for load balancer tagging.

## 0.5.0

* Allow load balancer functionality to be disabled. [#198][20]
* Allow subnets in configuration to be optional. [#202][21]

## 0.4.0

 * Implement `loadbalancer.securityListManagementMode: Frontend` which only
   manages security list rules for load balancer ingress. [#180][16]
 * Depreciate `loadbalancer.disableSecurityListManagement` in favour of
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
 * Depreciate cloud-provider config property `auth.key_passphrase` replacing it
   with `auth.passphrase` [#142][12]
 * Depreciate cloud-provider config property `auth.compartment` replacing it
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
[20]: https://github.com/oracle/oci-cloud-controller-manager/pull/199
[21]: https://github.com/oracle/oci-cloud-controller-manager/pull/204
[22]: https://github.com/oracle/oci-cloud-controller-manager/issues/213
[23]: https://github.com/oracle/oci-cloud-controller-manager/issues/217
[24]: https://github.com/oracle/oci-cloud-controller-manager/pull/223
[25]: https://github.com/oracle/oci-cloud-controller-manager/pull/228
[26]: https://github.com/oracle/oci-cloud-controller-manager/issues/108
[27]: https://github.com/oracle/oci-cloud-controller-manager/issues/232
[28]: https://github.com/oracle/oci-cloud-controller-manager/issues/208
[29]: https://github.com/oracle/oci-cloud-controller-manager/pull/238
[30]: https://github.com/oracle/oci-cloud-controller-manager/issues/235
