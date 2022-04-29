# Block Volume Expansion using CSI

## Setup 

1. Make sure you have installed [CCM](../README.md) and [CSI](../container-storage-interface.md) version v1.19.12 or later

To create a PVC backed by a block volume with a Lower Cost, Balanced, or Higher Performance performance level, set vpusPerGB in the storage class definition as follows:

* for a Lower Cost performance level, set vpusPerGB: "0"
* for a Balanced performance level, set vpusPerGB: "10"
* for a Higher Performance performance level, set vpusPerGB: "20"

## Create Storage Class for high performance
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: oci-high
provisioner: blockvolume.csi.oraclecloud.com
parameters:
  vpusPerGB: "20"
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

The value of vpusPerGB must be "0", "10", or "20". Other values are not supported.

## Create PVC

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: oci-pvc-high
spec:
  storageClassName: oci-high
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```

## Create POD

```yaml
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
        claimName: oci-pvc-high
```

For more information refer [CSI BV Performance Doc][1]

Note: 
Performance of block volume can be specified at the creation itself. Performance (vpusPerGB) cannot be modified after volume is provisioned.
CSI version 1.19.12 or later which runs on k8s cluster 1.19 or later supports block volume expansion.
Flex volume does not support. 

[1]: https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengcreatingpersistentvolumeclaim.htm#contengcreatingpersistentvolumeclaim_topic_Provisioning_PVCs_on_BV_PV_Volume_performance
