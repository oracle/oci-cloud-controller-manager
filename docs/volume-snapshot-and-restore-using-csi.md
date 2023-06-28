# Block Volume Snapshot and Restore using CSI

## Setup

1. Make sure you have installed [CCM](../README.md) and [CSI](../container-storage-interface.md) version v1.26.0 or later

A Kubernetes volume snapshot is a snapshot of a persistent volume on a storage system. You can use a volume snapshot to provision a new persistent volume. For more information, about Kubernetes volume snapshots, see [Volume Snapshots][1] in the Kubernetes documentation.

You can use the CSI volume plugin to provision a volume snapshot in one of two ways:

Dynamically: You can request the creation of a backup of the block volume provisioning a persistent volume. You specify the persistent volume claim using the VolumeSnapshot object, and you specify the parameters to use to create the block volume backup using the VolumeSnapshotClass object. See [Creating Dynamically Provisioned Volume Snapshots](#creating-dynamically-provisioned-volume-snapshots).

Statically: You can provide details of an existing block volume backup using the VolumeSnapshotContent object. See [Creating Statically Provisioned Volume Snapshots](#creating-statically-provisioned-volume-snapshots)

Note the following when creating and using volume snapshots:

* In the case of dynamic volume backups, the CSI volume plugin creates a new block volume backup to provision a dynamic volume snapshot in the same compartment as the cluster. In the case of static volume snapshots, the block volume backup provisioning a static volume snapshot can be in a different compartment to the cluster, provided appropriate policy statements exist to enable the cluster to access that other compartment (see [Prerequisites for Creating Volume Snapshots](#prerequisites-for-creating-volume-snapshots)).
* You cannot use the CSI volume plugin to re-populate an existing volume with data. In other words, you cannot restore (revert) data in an existing persistent volume to an earlier state by changing the volume snapshot specified in the persistent volume claim's manifest. You can only use the CSI volume plugin to populate a new volume with data.
* Cross-namespace snapshots are not supported.

### Prerequisites for Creating Volume Snapshots

* The VolumeSnapshot, VolumeSnapshotContent, and VolumeSnapshotClass objects are not part of the core Kubernetes API. Therefore, before you can create volume snapshots using the CSI volume plugin, you have to install the necessary CRD (Custom Resource Definition) files on the cluster, by running the following commands:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml
```

* If you want to use a statically provisioned volume snapshot to provision a new persistent volume, and the underlying block volume backup is in a different compartment to the cluster, appropriate policy statements must exist to enable the cluster to access the block volume backups in that other compartment. For example:

```bash
ALLOW any-user to manage volume-backups in compartment <compartment-name> where request.principal.type = 'cluster'
ALLOW any-user to use volumes in compartment <compartment-name> where request.principal.type = 'cluster'
```

## Creating Dynamically Provisioned Volume Snapshots

To dynamically provision a volume snapshot by creating a backup of the block volume provisioning a persistent volume claim, you first define a VolumeSnapshotClass object that specifies the type of block volume backup to create. Having created the VolumeSnapshotClass object, you then define a VolumeSnapshot object that uses the VolumeSnapshotClass. You use the VolumeSnapshot object to specify the persistent volume claim provisioned by the block volume that you want to back up.

For example, you define a persistent volume claim named sample-pvc in a file called csi-mypvctobackup.yaml, provisioned by a block volume:

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sample-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "oci-bv"
  resources:
    requests:
      storage: 50Gi
```
Create the persistent volume claim:

```
kubectl create -f csi-mypvctobackup.yaml
```

You can use the persistent volume claim when defining other objects, such as pods. For example, the following pod definition instructs the system to use the sample-pvc persistent volume claim as the nginx volume, which is mounted by the pod at /sample-volume.

```
apiVersion: v1
kind: Pod
metadata:
  name: sample-pod
spec:
  containers:
    - name: sample-nginx
      image: nginx
      ports:
        - containerPort: 80
          name: "http-server"
      volumeMounts:
        - mountPath: "/usr/share/nginx/html"
          name: sample-volume
  volumes:
  - name: sample-volume
    persistentVolumeClaim:
      claimName: sample-pvc
```

Having created the new pod, the persistent volume claim is bound to a new persistent volume provisioned by a block volume.

In readiness for creating a backup of the block volume provisioning the persistent volume claim, you set parameters for the block volume backup by defining a VolumeSnapshotClass object named my-snapclass in a file called csi-mysnapshotclass.yaml:

```
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: my-snapclass
driver: blockvolume.csi.oraclecloud.com
parameters:
  backupType: full
deletionPolicy: Delete
```

where:

* driver: blockvolume.csi.oraclecloud.com specifies the CSI volume plugin to provision VolumeSnapshot objects.
* parameters.backupType: full specifies a block volume backup is to include all changes since the block volume was created. Specify incremental to create a backup with only the changes since the last backup. Note that for data recovery purposes, there is no functional difference between an incremental backup and a full backup. See [Volume Backup Types][3].
* deletionPolicy: Delete specifies what happens to a block volume backup if the associated VolumeSnapshot object is deleted. Specify Retain to keep a block volume backup if the associated VolumeSnapshot object is deleted.

By default, the same freeform tags and defined tags that were applied to the source block volume are applied to the block volume backup. However, you can use annotations to apply additional tags to the block volume backup (see Tagging Block Volume Backups).

Create the VolumeSnapshotClass object:

```
kubectl create -f csi-mysnapshotclass.yaml
```

To create a backup of the block volume provisioning the persistent volume claim, you then define a VolumeSnapshot object as my-snapshot in a file called csi-mysnapshot.yaml:

```
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-snapshot
  namespace: default
spec:
  volumeSnapshotClassName: my-snapclass
  source:
    persistentVolumeClaimName: sample-pvc
```

where:

* volumeSnapshotClassName: my-snapclass specifies my-snapclass as the VolumeSnapshotClass object from which to obtain parameters to use when creating the block volume backup. Note that you cannot change volumeSnapshotClassName after you have created the VolumeSnapshot object (you have to create a new VolumeSnapshot object).
* persistentVolumeClaimName: sample-pvc specifies sample-pvc as the persistent volume claim based on the block volume for which you want to create a block volume backup. Note that you cannot change the source after you have created the VolumeSnapshot object (you have to create a new VolumeSnapshot object).

Create the VolumeSnapshot object:

```
kubectl create -f csi-mysnapshot.yaml
```

The VolumeSnapshot object is created and provisioned by a new block volume backup. You can use the volume snapshot to provision a new persistent volume (see [Using a Volume Snapshot to Provision a New Volume](#using-a-volume-snapshot-to-provision-a-new-volume)).

## Creating Statically Provisioned Volume Snapshots

To statically provision a volume snapshot from an existing block volume backup, you first create the block volume backup (see [Backing Up a Volume][2]).

Having created the block volume backup, define a VolumeSnapshotContent object and specify details (including the OCID) of the existing block volume backup. You can then define a VolumeSnapshot object and specify the VolumeSnapshotContent object that provides details of the existing block volume backup.

For example, you define the VolumeSnapshotContent object as my-static-snapshot-content in a file called csi-mystaticsnapshotcontent.yaml:

```
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotContent
metadata:
  name: my-static-snapshot-content
spec:
  deletionPolicy: Retain
  driver: blockvolume.csi.oraclecloud.com
  source:
    snapshotHandle: ocid1.volumebackup.oc1.iad.aaaaaa______xbd
  volumeSnapshotRef:
    name: my-static-snapshot
    namespace: default
```

where:
* deletionPolicy: Retain specifies what happens to a block volume backup if the associated VolumeSnapshot object is deleted. Specify Delete to delete a block volume backup if the associated VolumeSnapshot object is deleted.
* driver: blockvolume.csi.oraclecloud.com specifies to use the CSI volume plugin to provision VolumeSnapshot objects.
* snapshotHandle: ocid1.volumebackup.oc1.iad.aaaaaa______xbd specifies the OCID of the existing block volume backup.
* volumeSnapshotRef.name: my-static-snapshot specifies the name of the corresponding VolumeSnapshot object to be provisioned from the existing block volume backup. This field is required. Note that the VolumeSnapshot object need not exist when you create the VolumeSnapshotClass object.
* namespace: default specifies the namespace containing the corresponding VolumeSnapshot object to be provisioned from the existing block volume backup. This field is required.

Create the VolumeSnapshotClass object:

```
kubectl create -f csi-mystaticsnapshotcontent.yaml
```

You define the statically provisioned VolumeSnapshot object as my-static-snapshot in a file called csi-mystaticsnapshot.yaml:

```
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-static-snapshot
spec:
  source:
    volumeSnapshotContentName: static-snapshot-content
```

where VolumeSnapshotContentName: static-snapshot-content specifies the name of the VolumeSnapshotContent object you created previously. Note that you cannot change the source after you have created the VolumeSnapshot object (you have to create a new VolumeSnapshot object).

Create the VolumeSnapshot object:
```
kubectl create -f csi-mystaticsnapshot.yaml
```
The VolumeSnapshot object is created and provisioned by the block volume backup specified in the VolumeSnapshotContent object. You can use the volume snapshot to provision a new persistent volume (see [Using a Volume Snapshot to Provision a New Volume](#using-a-volume-snapshot-to-provision-a-new-volume)).

## Using a Volume Snapshot to Provision a New Volume

Having created a dynamically provisioned or statically provisioned volume snapshot, you can specify the volume snapshot as the datasource for a persistent volume claim to provision a new persistent volume.

For example, you define a persistent volume claim named pvc-fromsnapshot in a file called csi-mypvcfromsnapshot.yaml, provisioned by a volume snapshot named test-snapshot:

```
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-fromsnapshot
  namespace: default
spec:
  storageClassName: oci-bv
  dataSource:
    name: test-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```

where:

* datasource.name: test-snapshot specifies test-snapshot as the name of the VolumeSnapshot object to use as the data source for the persistent volume.
* datasource.apiGroup: snapshot.storage.k8s.io specifies the version of the Kubernetes snapshot storage API to use.

Create the persistent volume claim:
```
kubectl create -f csi-mypvcfromsnapshot.yaml
```
When the persistent volume claim is used to provision another object (such as a pod), a persistent volume is created and the VolumeSnapshot object you specified is used to populate the underlying block volume. For example, you could create a new pod from the following pod definition that instructs the system to use the pvc-fromsnapshot PVC as the nginx volume, which is mounted by the pod at /data.

```
apiVersion: v1
kind: Pod
metadata:
  name: sample-pod-restore
spec:
  containers:
    - name: nginx
      image: nginx:latest
      ports:
        - name: http
          containerPort: 80
      volumeMounts:
        - name: data
          mountPath: /data
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: pvc-fromsnapshot
```

Having created the new pod, the persistent volume claim is bound to a new persistent volume provisioned by a new block volume populated by the VolumeSnapshot object.

[1]: https://kubernetes.io/docs/concepts/storage/volume-snapshots/
[2]: https://docs.oracle.com/en-us/iaas/Content/Block/Tasks/backingupavolume.htm#Backing_Up_a_Volume
[3]: https://docs.oracle.com/en-us/iaas/Content/Block/Concepts/blockvolumebackups.htm#backuptype
