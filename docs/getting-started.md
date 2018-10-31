# Getting Started

cloud-provider-oci contains the following components for use when operating
Kubernetes clusters on Oracle Cloud Infrastructure (OCI):

 - `oci-cloud-controller-manager`: A Kubernetes [cloud controller manager][1]
   daemon that runs OCI cloud-specific control loops including node
   initialization and loadbalancer provisioning.
 - [`oci-volume-provisioner`][6]: A Kubenetes [dynamic volume provisioner][2] for
   provisioning OCI block storage and file system storage persistant volumes.
 - [`oci-flexvolume-driver`][7]: A Kubernetes [flexvolume driver][3] used in
   conjunction with the oci-volume-provisioner to attach provisioned volumes to
   Kubernetes Nodes.

## OCI Cloud Controller Mangager

### Requirements

At the time of writing running a Kubernetes cloud controller manager requires
a few things. Please read through the requirements carefully as they are
critical to running cloud controller manager for a Kubernetes cluster on OCI.

#### `--cloud-provider=external`

All `kubelet`s in your cluster **MUST** set the flag `--cloud-provider=external`.
`kube-apiserver` and `kube-controller-manager` must **NOT** set the
`--cloud-provider` flag which will default them to use no cloud provider
natively.

**WARNING**: setting `--cloud-provider=external` on the `kubelet` will taint all
Nodes in a cluster with `node.cloudprovider.kubernetes.io/uninitialized`. It is
then the responsibility of the cloud controller manager to untaint those nodes
once it has finished initializing them. This means that the majority of Pods
will be left unscheduable until the cloud controller manager is running.

In the future, `--cloud-provider=external` will be the default. Learn more about
the future of cloud providers in Kubernetes [here][4].

#### Node names must match the instance display name, primary vNIC hostnameLabel, or public IPv4 IP

By default, the `kubelet` will name Nodes based on the node's hostname. On
OCI, instance hostnames are set based on either the hostnameLabel of the
instance's primary VNIC (if provided) or the instance display name. If you
decide to override the hostname on kubelets with `--hostname-override`, this
will also override the Node name in Kubernetes.

It is important that the Node name in Kubernetes matches either the instance
name, primary VNIC hostname label, or the public IPv4 IP, otherwise the cloud
controller manager will not be able to find the corresponding instances for the
Kubernetes Nodes.

When setting the Instance host name as the node name (which is the default),
Kubernetes will try to reach the Node using its hostname. However, this won't
neccicarily work depending on your VCN setup since hostnames may not be
resolvable. For example, when you run `kubectl logs` you may get an error like
the following:

```
$ kubectl logs -f mypod
Error from server: Get https://k8s-worker-03:10250/containerLogs/default/mypod/mypod?follow=true: dial tcp: lookup k8s-worker-03 on 67.207.67.3:53: no such host
```

To prevent this it's important to tell the Kubernetes masters to use another
address type to reach its workers. You can do this by setting
`--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname` on the
`kube-apiserver`. Doing this will tell Kubernetes to use an instance's private
IP to connect to the node before attempting it's public IP and then it's
hostname.

#### All Instances must have unique names

The instance display names for all instances in your kubernetes cluster must be
unique as Node names in Kubernetes must be unique.

### Implementation Details

Currently `oci-cloud-controller-manager` implements:
 - NodeController - updates nodes with cloud provider specific labels and
   addresses, also deletes Nodes from Kubernetes when the corosponding instance
   is deleted from the cloud-provider.
 - ServiceController - responsible for creating load balancers when a Service
   of `type: LoadBalancer` is created in Kubernetes.

### Deployment

You will first need to create a cloud-provider configuration file tailored to
your OCI account and Kubernetes cluster. Please see the provider configuration
documentation [here][5].

```bash
$ kubectl  create secret generic oci-cloud-controller-manager \
     -n kube-system                                           \
     --from-file=cloud-provider.yaml=cloud-provider.yaml
```

```bash
kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/0.7.0/oci-cloud-controller-manager-rbac.yaml
kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/0.7.0/oci-cloud-controller-manager.yaml
```

NOTE: the release deployments are meant to serve as an example. They will work
in a majority of cases but may not work out of the box for your cluster.

[1]: https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/
[2]: https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/
[3]: https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md
[4]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/cloud-provider/cloud-provider-refactoring.md
[5]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/docs/provider-configuration.md
[6]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/docs/using-oci-volume-provisioner.md
[7]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/docs/using-oci-flexvolume-driver.md
