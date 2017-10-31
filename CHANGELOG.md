# CHANGELOG

## 0.1.0 (alpha) - 30th October 2017
* Implement `cloudprovider.Instances` interface - keeps track of node state within an OCI Kubernetes cluster.
* Implement `cloudprovider.LoadBalancer` interface - enables create, update, and delete of OCI load balancers for services of type `LoadBalancer`.
* Implement `cloudprovider.Zones` interface - provides OCI region information for cluster nodes.
