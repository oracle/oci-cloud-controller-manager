# Provisioning PVCs on the File Storage with Lustre Service

The Oracle Cloud Infrastructure File Storage with Lustre service is a fully managed storage service designed to meet the demands of AI/ML training and inference, and high performance computing needs. You use the Lustre CSI plugin to connect clusters to file systems in the File Storage with Lustre service.

You can use the File Storage with Lustre service to provision persistent volume claims (PVCs) by manually creating a file system in the File Storage with Lustre service, then defining and creating a persistent volume (PV) backed by the new file system, and finally defining a new PVC. When you create the PVC, Kubernetes binds the PVC to the PV backed by the File Storage with Lustre service.

The Lustre CSI driver is the overall software that enables Lustre file systems to be used with Kubernetes via the Container Storage Interface (CSI). The Lustre CSI plugin is a specific component within the driver, responsible for interacting with the Kubernetes API server and managing the lifecycle of Lustre volumes.

Note the following:

- The Lustre CSI driver is supported on Oracle Linux 8 x86 and on Ubuntu x86 22.04.
- To use a Lustre file system with a Kubernetes cluster, the Lustre client package must be installed on worker nodes that have to mount the file system. For more information about Lustre clients, see [Mounting and Accessing a Lustre File System](https://docs.oracle.com/iaas/Content/lustre/file-system-connect.htm).

## Provisioning a PVC on an Existing File System

To create a PVC on an existing file system in the File Storage with Lustre service (using Oracle-managed encryption keys to encrypt data at rest):

1. Create a file system in the File Storage with Lustre service, selecting the Encrypt using Oracle-managed keys encryption option. See [Creating a Lustre File System](https://docs.oracle.com/iaas/Content/lustre/file-system-create.htm).

2. Create security rules in either a network security group (recommended) or a security list for both the Lustre file system, and for the cluster's worker nodes subnet. The security rules to create depend on the relative network locations of the Lustre file system and the worker nodes which act as the client, according to the following scenarios:

   These scenarios, the security rules to create, and where to create them, are fully described in the File Storage with Lustre service documentation (see [Required VCN Security Rules](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm)).

3. Create a PV backed by the file system in the File Storage with Lustre service as follows:

   a. Create a manifest file to define a PV and in the `csi:` section, set:

    - `driver` to `lustre.csi.oraclecloud.com`
    - `volumeHandle` to `<MGSAddress>@<LNetName>:/<MountName>`
      where:
        - `<MGSAddress>` is the Management service address for the file system in the File Storage with Lustre service
        - `<LNetName>` is the LNet network name for the file system in the File Storage with Lustre service
        - `<MountName>` is the mount name used while creating the file system in the File Storage with Lustre service

      For example: `10.0.2.6@tcp:/testlustrefs`

    - `fsType` to `lustre`
    - (optional, but recommended) `volumeAttributes.setupLnet` to `"true"` if you want the Lustre CSI driver to perform lnet (Lustre Network) setup before mounting the filesystem
    - (required) `volumeAttributes.lustreSubnetCidr` to the CIDR block of the subnet where the worker node's VNIC having access to lustre filesystem is located (typically worker node subnet in default setup) to ensure the worker node has network connectivity to the Lustre file system. For example, 10.0.2.0/24.
    - (optional) `volumeAttributes.lustrePostMountParameters` to set Lustre parameters. For example:
      ```yaml
      volumeAttributes:
        lustrePostMountParameters: '[{"*.*.*MDT*.lru_size": 11200},{"at_history" : 600}]'
      ```

   For example, the following manifest file (named `lustre-pv-example.yaml`) defines a PV called `lustre-pv-example` backed by a Lustre file system:

   ```yaml
   apiVersion: v1
   kind: PersistentVolume
   metadata:
     name: lustre-pv-example
   spec:
     capacity:
       storage: 31Ti
     volumeMode: Filesystem
     accessModes:
     - ReadWriteMany
     persistentVolumeReclaimPolicy: Retain
     csi:
       driver: lustre.csi.oraclecloud.com
       volumeHandle: "10.0.2.6@tcp:/testlustrefs"
       fsType: lustre
       volumeAttributes:
         setupLnet: "true"
   ```

   b. Create the PV from the manifest file by entering:
   ```bash
   kubectl apply -f <filename>
   ```

   For example:
   ```bash
   kubectl apply -f lustre-pv-example.yaml
   ```

   c. Verify that the PV has been created successfully by entering:
   ```bash
   kubectl get pv <pv-name>
   ```

   For example:
   ```bash
   kubectl get pv lustre-pv-example
   ```

   Example output:
   ```
   NAME                CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM   STORAGECLASS   REASON   AGE
   lustre-pv-example   31Ti       RWX            Retain           Bound                                     56m
   ```

4. Create a PVC that is provisioned by the PV you have created, as follows:

   a. Create a manifest file to define the PVC and set:

    - `storageClassName` to `""`

      **Note:** You must specify an empty value for `storageClassName`, even though storage class is not applicable in the case of static provisioning of persistent storage. If you do not specify an empty value for `storageClassName`, the default storage class (`oci-bv`) is used, which causes an error.

    - `volumeName` to the name of the PV you created (for example, `lustre-pv-example`)

   For example, the following manifest file (named `lustre-pvc-example.yaml`) defines a PVC named `lustre-pvc-example` that will bind to a PV named `lustre-pv-example`:

   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: lustre-pvc-example
   spec:
     accessModes:
     - ReadWriteMany
     storageClassName: ""
     volumeName: lustre-pv-example
     resources:
       requests:
         storage: 31Ti
   ```

   **Note:** The `requests: storage:` element must be present in the PVC's manifest file, and its value must match the value specified for the `capacity: storage:` element in the PV's manifest file. Apart from that, the value of the `requests: storage:` element is ignored.

   b. Create the PVC from the manifest file by entering:
   ```bash
   kubectl apply -f <filename>
   ```

   For example:
   ```bash
   kubectl apply -f lustre-pvc-example.yaml
   ```

   c. Verify that the PVC has been created and bound to the PV successfully by entering:
   ```bash
   kubectl get pvc <pvc-name>
   ```

   For example:
   ```bash
   kubectl get pvc lustre-pvc-example
   ```

   Example output:
   ```
   NAME                  STATUS   VOLUME              CAPACITY   ACCESS MODES   STORAGECLASS   AGE
   lustre-pvc-example    Bound    lustre-pv-example   31Ti       RWX                           57m
   ```

   The PVC is bound to the PV backed by the File Storage with Lustre service file system. Data is encrypted at rest, using encryption keys managed by Oracle.

5. Use the new PVC when creating other objects, such as deployments. For example:

   a. Create a manifest named `lustre-app-example-deployment.yaml` to define a deployment named `lustre-app-example-deployment` that uses the `lustre-pvc-example` PVC, as follows:

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: lustre-app-example-deployment
   spec:
     selector:
       matchLabels:
         app: lustre-app-example
     replicas: 2
     template:
       metadata:
         labels:
           app: lustre-app-example
       spec:
         containers:
         - args:
           - -c
           - while true; do echo $(date -u) >> /lustre/data/out.txt; sleep 60; done
           command:
           - /bin/sh
           image: busybox:latest
           imagePullPolicy: Always
           name: lustre-app-example
           volumeMounts:
           - mountPath: /lustre/data
             name: lustre-volume
         restartPolicy: Always
         volumes:
         - name: lustre-volume
           persistentVolumeClaim:
             claimName: lustre-pvc-example
   ```

   b. Create the deployment from the manifest file by entering:
   ```bash
   kubectl apply -f lustre-app-example-deployment.yaml
   ```

   c. Verify that the deployment pods have been created successfully and are running by entering:
   ```bash
   kubectl get pods
   ```

   Example output:
   ```
   NAME                                           READY   STATUS    RESTARTS   AGE
   lustre-app-example-deployment-7767fdff86-nd75n   1/1     Running   0          8h
   lustre-app-example-deployment-7767fdff86-wmxlh   1/1     Running   0          8h
   ```

## Provisioning a PVC on an Existing File System with Mount Options

You can optimize the performance and control access to an existing Lustre file system by specifying mount options for the PV. Specifying mount options enables you to fine-tune how pods interact with the file system.

To include mount options:

1. Start by following the instructions in [Provisioning a PVC on an Existing File System](#provisioning-a-pvc-on-an-existing-file-system).

2. In the PV manifest described in [Provisioning a PVC on an Existing File System](#provisioning-a-pvc-on-an-existing-file-system), add the `spec.mountOptions` field, which enables you to specify how the PV should be mounted by pods.

   For example, in the `lustre-pv-example.yaml` manifest file shown in [Provisioning a PVC on an Existing File System](#provisioning-a-pvc-on-an-existing-file-system), you can include the `mountOptions` field as follows:

   ```yaml
   apiVersion: v1
   kind: PersistentVolume
   metadata:
     name: lustre-pv-example
   spec:
     capacity:
       storage: 31Ti
     volumeMode: Filesystem
     accessModes:
     - ReadWriteMany
     persistentVolumeReclaimPolicy: Retain
     mountOptions:
     - ro
     csi:
       driver: lustre.csi.oraclecloud.com
       volumeHandle: "10.0.2.6@tcp:/testlustrefs"
       fsType: lustre
       volumeAttributes:
         setupLnet: "true"
   ```

   In this example, the `mountOptions` field is set to `ro`, indicating that pods are to have read-only access to the file system. For more information about PV mount options, see [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) in the Kubernetes documentation.

## Encrypting Data At Rest on an Existing File System

The File Storage with Lustre service always encrypts data at rest, using Oracle-managed encryption keys by default. However, you have the option to specify at-rest encryption using your own master encryption keys that you manage yourself in the Vault service.

For more information about creating File Storage with Lustre file systems that use Oracle-managed encryption keys or your own master encryption keys that you manage yourself, see [Updating File System Encryption](https://docs.oracle.com/iaas/Content/lustre/file-system-encryption.htm).
