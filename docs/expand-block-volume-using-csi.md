# Block Volume Expansion using CSI

## Setup 

1. Make sure you have installed [CCM](../README.md) and [CSI](../container-storage-interface.md) version v0.13.0 or later

## Create PVC

```yaml
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
        claimName: oci-bv-claim
```

Once PVC is Bound and Pod is running, edit the PVC to increase size of storage

Note: 
Only increasing the size of Block volume is supported, shrinking is not. 
CSI version 0.13.0 or later which runs on k8s cluster 1.19 or later supports block volume expansion.
Flex volume does not support. 
