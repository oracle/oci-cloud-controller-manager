# Container storage interface

This project implements a csi plugin for Kubernetes clusters
running on Oracle Cloud Infrastructure (OCI). It enables provisioning, attach, detach, mount and unmount of [OCI block
storage volumes][1] to Kubernetes Pods via the [CSI][2] plugin interface.

## Install / Setup

We publish two binaries oci-csi-controller-driver which runs on the master nodes, and oci-csi-node-controller which runs on each of the worker nodes.

### Submit configuration as a Kubernetes secret

Create a config.yaml file with contents similar to the following. This file will contain authentication
information necessary to authenticate with the OCI APIs and provision block storage volumes.
An example configuration file can be found [here](manifests/provider-config-example.yaml)

Submit this as a Kubernetes Secret.

```bash
kubectl create secret generic oci-volume-provisioner \
    -n kube-system \
    --from-file=config.yaml=provider-config-example.yaml
```
### Installer

Create the associated RBAC rules if your cluster is configured to use [RBAC][3]

Before applying the below yaml configs make sure to set the version you want to install
```bash
$ export RELEASE=?
```

```bash
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-csi-node-rbac.yaml
```
Deploy the csi-controller-driver:
It is provided as a deployment and it has three containers - 
    1. csi-provisioner [external-provisioner][4]
    2. csi-attacher [external-attacher][5]
    3. oci-csi-controller-driver

```bash
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-csi-controller-driver.yaml
```

Deploy the node-driver:
It is provided as a daemon set and it has two containers - 
    1. node-driver-registrar [node-driver-registrar][6]
    2. oci-csi-node-driver

```bash
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-csi-node-driver.yaml
```

Deploy the csi storage class:

```
$ kubectl apply -f https://raw.githubusercontent.com/oracle/oci-cloud-controller-manager/master/manifests/container-storage-interface/storage-class.yaml
```

Lastly, verify that the oci-csi-controller-driver and oci-csi-node-controller is running in your cluster. By default it runs in the 'kube-system' namespace.

```
$ kubectl -n kube-system get po | grep csi-oci-controller
$ kubectl -n kube-system get po | grep csi-oci-node
```

## Tutorial

Create a claim:

```bash
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: oci-bv-claim
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: oci-bv
  resources:
    requests:
      storage: 50Gi
```

Check PVC is in pending state:

```
$ kubectl describe pvc/oci-bv-claim
```

Create pod:

```bash
apiVersion: v1
kind: Pod
metadata:
  name: app1
spec:
  containers:
    - name: app1
      image: centos
      command: ["/bin/sh"]
      args: ["-c", "while true; do echo $(date -u) >> /data/out.txt; sleep 5; done"]
      volumeMounts:
        - name: persistent-storage
          mountPath: /data
  volumes:
    - name: persistent-storage
      persistentVolumeClaim:
        claimName: oci-bv-claim
``` 

Check if PVC is now in bound state:

```
$ kubectl describe pvc/oci-bv-claim
```

[1]: https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm
[2]: https://kubernetes.io/blog/2019/01/15/container-storage-interface-ga/
[3]: https://kubernetes.io/docs/admin/authorization/rbac/
[4]: https://kubernetes-csi.github.io/docs/external-provisioner.html
[5]: https://kubernetes-csi.github.io/docs/external-attacher.html
[6]: https://kubernetes-csi.github.io/docs/node-driver-registrar.html
