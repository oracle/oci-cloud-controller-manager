# Statically Provisioned FSS using CSI

## Setup 

1. Make sure you have installed [CCM](../README.md) and [CSI](../container-storage-interface.md) version v0.13.0 or later
2. Setup the [network configuration for FSS][1]
3. Create a [File System with Mount Target][2]

## Create PV

```bash
apiVersion: v1
kind: PersistentVolume
metadata:
  name: fss-static-pv
spec:
  capacity:
    storage: 50Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  csi:
    driver: fss.csi.oraclecloud.com
    volumeHandle: FileSystemOCID:serverIP:path
```

## Create PVC

```bash

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: fss-claim
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: ""
  resources:
    requests:
      storage: 5Gi
  volumeName: fss-static-pv
```

## Create POD

```bash
apiVersion: v1
kind: Pod
metadata:
  name: app
spec:
  containers:
  - name: app
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo $(date -u) >> /data/out.txt; sleep 5; done"]
    volumeMounts:
    - name: persistent-storage
      mountPath: /data
  volumes:
  - name: persistent-storage
    persistentVolumeClaim:
      claimName: fss-claim
```

# Statically Provisioned FSS with in-transit encryption using CSI

## Additional Setup for in-transit encryption

1. The statically provisioned FSS in-transit encryption tests need the oci-fss-utils package to be installed on the nodes running the tests. The package is needed to be downloaded from [here][3] and installed on the applicable nodes.
2. Additional [network configuration for in-transit encryption][4]

## Create PV

```bash
apiVersion: v1
kind: PersistentVolume
metadata:
  name: fss-in-transit-encrypted-static-pv
spec:
  capacity:
    storage: 50Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  csi:
    driver: fss.csi.oraclecloud.com
    volumeHandle: FileSystemOCID:serverIP:path
    volumeAttributes:
      encryptInTransit: "true"
```

## Create PVC

```bash

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: fss-claim
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: ""
  resources:
    requests:
      storage: 5Gi
  volumeName: fss-in-transit-encrypted-static-pv
```

## Create POD

```bash
apiVersion: v1
kind: Pod
metadata:
  name: app
spec:
  containers:
  - name: app
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo $(date -u) >> /data/out.txt; sleep 5; done"]
    volumeMounts:
    - name: persistent-storage
      mountPath: /data
  volumes:
  - name: persistent-storage
    persistentVolumeClaim:
      claimName: fss-claim
```

[1]: https://docs.oracle.com/en-us/iaas/Content/File/Tasks/securitylistsfilestorage.htm
[2]: https://docs.oracle.com/en-us/iaas/Content/File/Tasks/creatingfilesystems.htm
[3]: https://www.oracle.com/downloads/cloud/cloud-infrastructure-file-storage-downloads.html
[4]: https://docs.oracle.com/en-us/iaas/Content/File/Tasks/intransitencryption.htm#intransitencryption-prerequisites
