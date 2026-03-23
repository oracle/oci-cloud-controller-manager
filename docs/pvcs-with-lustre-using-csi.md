# Provisioning PVCs on the File Storage with Lustre Service
*Find out how to provision persistent volume claims for clusters you've created using Kubernetes Engine \(OKE\) by mounting file systems from the File Storage with Lustre service.*
The Oracle Cloud Infrastructure File Storage with Lustre service is a fully managed storage service designed to meet the demands of AI/ML training and inference, and high performance computing needs. You use the Lustre CSI plugin to connect clusters to file systems in the File Storage with Lustre service.
You can use the File Storage with Lustre service to provision persistent volume claims \(PVCs\) in two ways:
-   By defining and creating a new storage class, and then defining and creating a PVC referencing that storage class. When you create the PVC, the Lustre CSI plugin dynamically creates both a new Lustre file system and a new persistent volume backed by the new file system. See [Provisioning a PVC on a New Lustre File System Using the CSI Volume Plugin](#provisioning-a-pvc-on-a-new-lustre-file-system-using-the-csi-volume-plugin).
-   By manually creating a file system in the File Storage with Lustre service, then defining and creating a persistent volume \(PV\) backed by the new file system, and finally defining a new PVC. When you create the PVC, Kubernetes Engine binds the PVC to the PV backed by the File Storage with Lustre service. See [Provisioning a PVC on an Existing Lustre File System](#provisioning-a-pvc-on-an-existing-lustre-file-system)
    The Lustre CSI driver is the overall software that enables Lustre file systems to be used with Kubernetes via the Container Storage Interface \(CSI\). The Lustre CSI plugin is a specific component within the driver, responsible for interacting with the Kubernetes API server and managing the lifecycle of Lustre volumes.
    Note the following:
-   When using the Lustre CSI plugin to dynamically create a new file system, do not manually update or delete the persistent volume or Lustre file system objects that the CSI plugin creates.
-   Any Lustre file systems dynamically created by the CSI volume plugin are given names starting with `csi-lustre-` .
-   Any Lustre file systems dynamically created by the CSI volume plugin appear in the Console. However, do not use the Console \(or the Oracle Cloud Infrastructure CLI or API\) to modify these dynamically created resources. Changes made to Oracle Cloud Infrastructure resources dynamically created by the CSI volume plugin are not reconciled with Kubernetes objects.
-   Some advanced features such as volume expansion, snapshot, and clone are not currently available for dynamically provisioned Lustre file systems.
-   If you delete a PVC bound to a PV backed by a file system created by the CSI volume plugin and the reclaim policy is set to Delete, both the PV and the Lustre file system are deleted. If the reclaim policy is Retain, the PV is not deleted.
-   Using the Lustre CSI driver to provision a PVC on a dynamically created Lustre file system is supported on clusters created by Kubernetes Engine that are running Kubernetes version 1.32 or later. Using the Lustre CSI driver to provision a PVC on an existing Lustre file system is supported on clusters created by Kubernetes Engine that are running Kubernetes version 1.29 or later.
-   The Lustre CSI driver is supported on Oracle Linux 8 x86 and on Ubuntu x86 22.04.
-   To use a Lustre file system with a cluster created by Kubernetes Engine, the Lustre client package must be installed on worker nodes that have to mount the file system. For more information about Lustre clients, see [Mounting and Accessing a Lustre File System](https://docs.oracle.com/iaas/Content/lustre/file-system-connect.htm).
-   Data is encrypted at rest, using encryption keys managed either by Oracle or by you.
-   Oracle Cloud Infrastructure File Storage with Lustre is only available in the regions shown in [Availability](https://docs.oracle.com/iaas/Content/lustre/overview.htm#region-availability) in the File Storage with Lustre documentation.

## Provisioning a PVC on a New Lustre File System Using the CSI Volume Plugin
**Note:** The following prerequisites apply when provisioning a PVC on a new Lustre file system dynamically created by the CSI volume plugin:
-   Clusters must be running Kubernetes 1.32 or later to provision a PVC on a new file system dynamically created by the CSI volume plugin.
-   Appropriate IAM policies must exist to enable the CSI volume plugin to create and manage Lustre resources. For example:
    ```
    allow dynamic-group [your dynamic group name] to manage lustre-file-family in compartment <compartment-name>
    allow dynamic-group [your dynamic group name] to use virtual-network-family in compartment <compartment-name>
    ```
-   If the compartment to which a node pool, subnet, or file system belongs, is different to the compartment to which a cluster belongs, IAM policies must exist to enable the CSI volume plugin to access the appropriate location. For example:
    ```copy
    allow dynamic-group [your dynamic group name] to manage lustre-file-family in TENANCY
    ```
    ```copy
    allow dynamic-group [your dynamic group name] to use virtual-network-family in TENANCY
    ```
-   To specify a particular user-managed master encryption key from the [Vault](https://docs.oracle.com/iaas/Content/KeyManagement/Concepts/keyoverview.htm) service to encrypt data at rest, appropriate IAM policies must exist to enable the File Storage with Lustre service to access that master encryption key. See [Updating File System Encryption](https://docs.oracle.com/iaas/Content/lustre/file-system-encryption.htm).
-   The Lustre client package must be installed on all worker nodes that need to mount the Lustre file system.
    To dynamically provision a PVC on a new Lustre file system dynamically created by the CSI volume plugin in the File Storage with Lustre service:
1.  Create security rules in either a network security group \(recommended\) or a security list for both the Lustre file system, and for the cluster's worker nodes subnet.
    The security rules to create depend on the relative network locations of the Lustre file system and the worker nodes which act as the client, according to the following scenarios:
    -   [Option 1: Client and Lustre in Different Subnets](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm#top__different-subnet)
    -   [Option 2: Client and Lustre in the Same Subnet](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm#top__same-subnet)
        These scenarios, the security rules to create, and where to create them, are fully described in the File Storage with Lustre service documentation \(see [Required VCN Security Rules](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm)\).
2.  Define a new storage class that uses the `lustre.csi.oraclecloud.com` provisioner:
    1.  Create a manifest file \(for example, in a file named lustre-dyn-st-class.yaml\), specify a name for the new storage class, and specify values for required and optional parameters:
        ```
        kind: StorageClass
        apiVersion: storage.k8s.io/v1
        metadata:
          name: <storage-class-name>
        provisioner: lustre.csi.oraclecloud.com
        parameters:
          availabilityDomain: <ad-name>
          compartmentId: <compartment-ocid>   # optional
          subnetId: <subnet-ocid>
          performanceTier: <value>
          fileSystemName: <name>               # optional
          kmsKeyId: <key-ocid>                 # optional
          nsgIds: '["<nsg-ocid>"]'             # optional
          rootSquashEnabled: "<true | false>"  # optional
          rootSquashUid: "<value>"             # optional
          rootSquashGid: "<value>"             # optional
          rootSquashClientExceptions: '["<ip-address>"]'   # optional
          oci.oraclecloud.com/initial-defined-tags-override: '{"<tag-namespace>": {"<tag-key>": "<tag-value>"}}'
          oci.oraclecloud.com/initial-freeform-tags-override: '{"<tag-key>": "<tag-value>"}'
          setupLnet: "<true | false>"                    # optional
          lustreSubnetCidr: "<cidr-block>"      # optional
          lustrePostMountParameters: '[{"<parameter1>": <value>},{"<parameter2>": <value>}]' # optional
        ```
        where:
        -   `name: <storage-class-name>`: Required. A name of your choice for the storage class.
        -   `availabilityDomain: <ad-name>`: Required. The name of the availability domain in which to create the new Lustre file system. For example, `availabilityDomain: US-ASHBURN-AD-1`. To find out the availability domain name to use, run the `oci iam availability-domain list` CLI command \(or use the [ListAvailabilityDomains](/iaas/api/#/en/identity/latest/AvailabilityDomain/ListAvailabilityDomains) operation\). For more information, see [Your Tenancy's Availability Domain Names](https://docs.oracle.com/iaas/Content/General/Concepts/regions.htm#ad-names).
        -   `compartmentId: <compartment-ocid>`: Optional. The OCID of the compartment to which the new Lustre file system is to belong. If not specified, defaults to the same compartment as the cluster. For example, `compartmentId: ocid1.compartment.oc1..aaa______t6q`
        -   `subnetId: <subnet-ocid>`: Required. The OCID of the subnet in which to mount the new Lustre file system. For example, `subnetId: ocid1.subnet.oc1.iad.aaaa______kfa`
        -   `performanceTier: <value>`: Required. The Lustre file system performance tier. Allowed values:
            -   `MBPS_PER_TB_125`
            -   `MBPS_PER_TB_250`
            -   `MBPS_PER_TB_500`
            -   `MBPS_PER_TB_1000`
        -   `fileSystemName: <name>`: Optional. The Lustre file system name, up to 8 characters. If not specified, a default value will be randomly generated and used. For example, `fileSystemName: aiworkfs`
        -   `kmsKeyId: <key-ocid>`: Optional. The OCID of a master encryption key that you manage, with which to encrypt data at rest. If not specified, data is encrypted at rest using encryption keys managed by Oracle. For example, `kmsKeyId: ocid1.key.oc1.iad.ann______usj`
        -   `nsgIds: '["<nsg-ocid>"]'`: Optional. A JSON array of up to five network security group OCIDs to associate with the Lustre file system. For example, `nsgIds: '["ocid1.nsg.oc1.iad.aab______fea"]'`
        -   `rootSquashEnabled: "<true | false>"`: Optional. Set to `true` to restrict root access from clients. Defaults to `false`.
        -   `rootSquashUid: "<value>"`: Optional. When root squash is enabled, root operations are mapped to this UID. Defaults to `65534.`
        -   `rootSquashGid: "<value>"`: Optional. When root squash is enabled, root operations are mapped to this GID. Defaults to `65534`.
        -   `rootSquashClientExceptions: '["<ip-address>"]'`: Optional. A JSON array of IP addresses or CIDR blocks that are not subject to root squash \(maximum 10 entries\). For example, `rootSquashClientExceptions: '["10.0.2.4"]'`.
        -   `oci.oraclecloud.com/initial-defined-tags-override: '{"<tag-namespace>": {"<tag-key>": "<tag-value>"}}'` Optional. Specifies a defined tag for the new file system. For example, `oci.oraclecloud.com/initial-defined-tags-override: '{"Org": {"CostCenter": "AI"}}'`
            Note that to apply defined tags from a tag namespace belonging to one compartment to a filesystem resource belonging to a different compartment, you must include a policy statement to allow the cluster to use the tag namespace. See [Additional IAM Policy when a Cluster and a Tag Namespace are in Different Compartments](contengtaggingclusterresources_iam-tag-namespace-policy.md).
        -   `oci.oraclecloud.com/initial-freeform-tags-override: '{"<tag-key>": "<tag-value>"}'` Optional. Specifies a free-form tag for the new file system. For example, `oci.oraclecloud.com/initial-freeform-tags-override: '{"Project": "ML"}'`
        -   `setupLnet: "<true | false>"`: Optional. Set to `true` if the Lustre CSI driver should perform Lustre Network \(LNet\) setup before mounting. We strongly recommend you include the `setupLnet` parameter and set it `"true"`.
        -   `lustreSubnetCidr: "<cidr-block>"`: Optional. Set to the worker node's source network range used for Lustre traffic:
            -   **When to use:** Only specify a network range if worker nodes use a secondary VNIC to connect to the Lustre file system. This CIDR must match the subnet block of that secondary VNIC \(for example, `10.0.2.0/24`\).
            -   **When to omit:** Do not specify a network range if worker nodes are using their primary VNIC \(the default interface\) for Lustre connectivity.
            -   **Important:** This parameter is different to the Lustre file system's `subnetId` parameter, which defines where the Lustre file system itself is located.
        -   `lustrePostMountParameters: '[{"<parameter1>": <value>},{"<parameter2>": <value>}]'`: Optional. JSON array of advanced Lustre client parameters to set after mounting. For example, `lustrePostMountParameters: '[{"*.*.*MDT*.lru_size": 11200},{"at_history": 600}]'`
            For example:
        ```
        kind: StorageClass
        apiVersion: storage.k8s.io/v1
        metadata:
          name: lustre-dyn-storage
        provisioner: lustre.csi.oraclecloud.com
        parameters:
          availabilityDomain: US-ASHBURN-AD-1
          compartmentId: ocid1.compartment.oc1..aaa______t6q # optional
          subnetId: ocid1.subnet.oc1.iad.aaaa______kfa
          performanceTier: MBPS_PER_TB_250
          fileSystemName: aiworkfs                           # optional
          kmsKeyId: ocid1.key.oc1.iad.ann______usj           # optional
          nsgIds: '["ocid1.nsg.oc1.iad.aab______fea"]'       # optional
          oci.oraclecloud.com/initial-defined-tags-override: '{"Org": {"CostCenter": "AI"}}'
          oci.oraclecloud.com/initial-freeform-tags-override: '{"Project": "ML"}'
          setupLnet: "true"                    # optional
        ```
    2.  Create the storage class from the manifest file by entering:
        ```
        kubectl create -f <filename>
        ```
        For example:
        ```
        kubectl create -f lustre-dyn-st-class.yaml
        ```
3.  Create a PVC to be provisioned by the new file system in the File Storage with Lustre service, as follows:
    1.  Create a manifest file to define the PVC:
        ```
        apiVersion: v1
        kind: PersistentVolumeClaim
        metadata:
          name: <pvc-name>
        spec:
          accessModes:
            - <ReadWriteMany|ReadOnlyOncePod>
          storageClassName: "<storage-class-name>"
          resources:
            requests:
              storage: <capacity>
        ```
        where:
        -   `name: <pvc-name>`: Required. For example, `lustre-dynamic-claim`
        -   `storageClassName: "<storage-class-name>"`: Required. The name of the storage class you defined earlier. For example, `lustre-dyn-storage`.
        -   `accessModes: - <ReadWriteMany|ReadOnlyOncePod>`: Required. Specifies how the file system is to be mounted and shared by pods. Currently `ReadWriteMany` and `ReadOnlyOncePod` are supported. For example, `ReadWriteMany`.
        -   `storage: <capacity>`: Required. This value must be at least `31.2T` \(or `31200G`\). You can specify a larger capacity, but you must use particular increments that depend on capacity \(see [Increasing File System Capacity](https://docs.oracle.com/iaas/Content/lustre/file-system-capacity.htm)\). For example, `31.2T`.
            For example, the following manifest file \(named `lustre-dyn-claim.yaml`\) defines a PVC named `lustre-dynamic-claim` that is provisioned by the file system defined in the `lustre-dyn-storage` storage class:
        ```
        apiVersion: v1
        kind: PersistentVolumeClaim
        metadata:
          name: lustre-dynamic-claim
        spec:
          accessModes:
            - ReadWriteMany
          storageClassName: "lustre-dyn-storage"
          resources:
            requests:
              storage: 31.2T
        ```
    2.  Create the PVC from the manifest file by entering:
        ```
        kubectl create -f <filename> 
        ```
        For example:
        ```
        kubectl create -f lustre-dyn-claim.yaml
        ```
    A new PVC is created. The CSI volume plugin creates a new persistent volume \(PV\) and a new file system in the File Storage with Lustre service. Note that creating a new Lustre file system takes at least 10 minutes and can take longer, depending on the size of the file system. Use the Console or the CLI to confirm that the new Lustre file system has been created \(see [Listing File Systems](https://docs.oracle.com/iaas/Content/lustre/file-system-list.htm)\).
    The new PVC is bound to the new PV. Data is encrypted at rest, using encryption keys managed either by Oracle or by you.
4.  Verify that the PVC has been bound to the new persistent volume by entering:
    ```cloudshell
    kubectl get pvc
    ```
    Example output:
    ```scrollcopy
              
    NAME                   STATUS    VOLUME                    CAPACITY         ACCESSMODES   STORAGECLASS         AGE
    lustre-dynamic-claim   Bound     csi-lustre-<unique_ID>    30468750000Ki    RWX           lustre-dyn-storage   4m
    ```
5.  Use the new PVC when creating other objects, such as pods. For example, you could create a new pod from the following pod definition:
    ```
    apiVersion: v1
    kind: Pod
    metadata:
      name: lustre-dynamic-app
    spec:
      containers:
        - name: aiworkload
          image: busybox:latest
          command: ["sleep", "3600"]
          volumeMounts:
            - name: lustre-vol
              mountPath: /mnt/lustre
      volumes:
        - name: lustre-vol
          persistentVolumeClaim:
            claimName: lustre-dynamic-claim
    ```
6.  Having created a new pod as described in the example in the previous step, you can verify that the pod is using the new PVC by entering:
    ```
    kubectl describe pod lustre-dynamic-app
    ```

**Tip:**
If you foresee a frequent requirement to dynamically create new PVs and new filesystems when you create PVCs, you can specify that the new storage class you've created is to be used as the default storage class for provisioning new PVCs. See the [Kubernetes documentation](https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/) for more details.

### Encrypting Data At Rest on a New Lustre File System
The File Storage with Lustre service always encrypts data at rest, using Oracle-managed encryption keys by default. However, you have the option to specify at-rest encryption using your own master encryption keys that you manage yourself in the Vault service.
Depending on how you want to encrypt data at rest, follow the appropriate instructions below:
-   To use the CSI volume plugin to dynamically create a new Lustre file system that uses Oracle-managed encryption keys to encrypt data at rest, follow the steps in [Provisioning a PVC on a New Lustre File System Using the CSI Volume Plugin](#provisioning-a-pvc-on-a-new-lustre-file-system-using-the-csi-volume-plugin) and do not include the `kmsKeyId: <key-ocid>` parameter in the storage class definition. Data is encrypted at rest, using encryption keys managed by Oracle.
-   To use the CSI volume plugin to dynamically create a new Lustre file system that uses master encryption keys that you manage to encrypt data at rest, follow the steps in [Provisioning a PVC on a New Lustre File System Using the CSI Volume Plugin](#provisioning-a-pvc-on-a-new-lustre-file-system-using-the-csi-volume-plugin), include the `kmsKeyId: <key-ocid>` parameter in the storage class definition, and specify the OCID of the master encryption key in the Vault service. Data is encrypted at rest, using the encryption key you specify.

## Provisioning a PVC on an Existing Lustre File System
To create a PVC on an existing file system in the File Storage with Lustre service \(using Oracle-managed encryption keys to encrypt data at rest\):
1.  Create a file system in the File Storage with Lustre service, selecting the **Encrypt using Oracle-managed keys** encryption option. See [Creating a Lustre File System](https://docs.oracle.com/iaas/Content/lustre/file-system-create.htm).
2.  Create security rules in either a network security group \(recommended\) or a security list for both the Lustre file system, and for the cluster's worker nodes subnet.
    The security rules to create depend on the relative network locations of the Lustre file system and the worker nodes which act as the client, according to the following scenarios:
    -   [Option 1: Client and Lustre in Different Subnets](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm#top__different-subnet)
    -   [Option 2: Client and Lustre in the Same Subnet](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm#top__same-subnet)
        These scenarios, the security rules to create, and where to create them, are fully described in the File Storage with Lustre service documentation \(see [Required VCN Security Rules](https://docs.oracle.com/iaas/Content/lustre/security-rules.htm)\).
3.  Create a PV backed by the file system in the File Storage with Lustre service as follows:
    1.  Create a manifest file to define a PV and in the `csi:` section, set:
        -   `driver` to `lustre.csi.oraclecloud.com`
        -   `volumeHandle` to `<MGSAddress>@<LNetName>:/<MountName>`
            where:
            -   `<MGSAddress>` is the Management service address for the file system in the File Storage with Lustre service
            -   `<LNetName>` is the LNet network name for the file system in the File Storage with Lustre service.
            -   `<MountName>` is the mount name used while creating the file system in the File Storage with Lustre service.
                For example: `10.0.2.6@tcp:/testlustrefs`
        -   `fsType` to `lustre`
        -   \(optional, but recommended\) `volumeAttributes.setupLnet` to `"true"` if you want the Lustre CSI driver to perform lnet \(Lustre Network\) setup before mounting the filesystem.
        -   \(optional\) `volumeAttributes.lustreSubnetCidr` to the worker node's source network range used for Lustre traffic:
            -   **When to use:** Only specify a network range if worker nodes use a secondary VNIC to connect to the Lustre file system. This CIDR must match the subnet block of that secondary VNIC \(for example, `10.0.2.0/24`\).
            -   **When to omit:** Do not specify a network range if worker nodes are using their primary VNIC \(the default interface\) for Lustre connectivity.
            -   **Important:** This parameter is different to the Lustre file system's `subnetId` parameter, which defines where the Lustre file system itself is located.
        -   \(optional\) `volumeAttributes.lustrePostMountParameters` to set Lustre parameters. For example:
            ```
            ...
                volumeAttributes:
                  lustrePostMountParameters: '[{"*.*.*MDT*.lru_size": 11200},{"at_history" :
                600}]'
            ```
        For example, the following manifest file \(named lustre-pv-example.yaml\) defines a PV called `lustre-pv-example` backed by a Lustre file system:
        ```
        apiVersion: v1
        kind: PersistentVolume
        metadata:
          name: lustre-pv-example
        spec:
          capacity:
            storage: 31.2T
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
    2.  Create the PV from the manifest file by entering:
        ```
        kubectl apply -f <filename>
        ```
        For example:
        ```
        kubectl apply -f lustre-pv-example.yaml
        ```
    3.  Verify that the PV has been created successfully by entering:
        ```
        kubectl get pv <pv-name>
        ```
        For example:
        ```
        kubectl get pv lustre-pv-example
        ```
        Example output:
        ```
        
        NAME                CAPACITY        ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM   STORAGECLASS   REASON   AGE
        lustre-pv-example   30468750000Ki   RWX            Retain           Bound                                    56m
        
        ```
4.  Create a PVC that is provisioned by the PV you have created, as follows:
    1.  Create a manifest file to define the PVC and set:
        -   `storageClassName` to `""` Note that you must specify an empty value for `storageClassName`, even though storage class is not applicable in the case of static provisioning of persistent storage. If you do not specify an empty value for `storageClassName`, the default storage class \(`oci-bv`\) is used, which causes an error.
        -   `volumeName` to the name of the PV you created \(for example, `lustre-pv-example`\)
            For example, the following manifest file \(named lustre-pvc-example.yaml\) defines a PVC named `lustre-pvc-example` that will bind to a PV named `lustre-pv-example`:
        ```
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
              storage:  31.2T
        ```
        Note that the `requests: storage:` element must be present in the PVC's manifest file, and its value must match the value specified for the `capacity: storage:` element in the PV's manifest file. Apart from that, the value of the `requests: storage:` element is ignored.
    2.  Create the PVC from the manifest file by entering:
        ```
        kubectl apply -f <filename>
        ```
        For example:
        ```
        kubectl apply -f lustre-pvc-example.yaml
        ```
    3.  Verify that the PVC has been created and bound to the PV successfully by entering:
        ```
        kubectl get pvc <pvc-name>
        ```
        For example:
        ```
        kubectl get pvc lustre-pvc-example
        ```
        Example output:
        ```
        
        NAME                    STATUS   VOLUME              CAPACITY         ACCESS MODES   STORAGECLASS   AGE
        lustre-pvc-example      Bound    lustre-pv-example   30468750000Ki    RWX                           57m
        ```
    The PVC is bound to the PV backed by the File Storage with Lustre service file system. Data is encrypted at rest, using encryption keys managed by Oracle.
5.  Use the new PVC when creating other objects, such as deployments. For example:
    1.  Create a manifest named lustre-app-example-deployment.yaml to define a deployment named `lustre-app-example-deployment` that uses the `lustre-pvc-example` PVC, as follows:
        ```
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
    2.  Create the deployment from the manifest file by entering:
        ```
        kubectl apply -f lustre-app-example-deployment.yaml
        ```
    3.  Verify that the deployment pods have been created successfully and are running by entering:
        ```
        kubectl get pods
        ```
        Example output:
        ```
        NAME                                             READY   STATUS              RESTARTS   AGE
        lustre-app-example-deployment-7767fdff86-nd75n   1/1     Running             0          8h
        lustre-app-example-deployment-7767fdff86-wmxlh   1/1     Running             0          8h
        ```

### Provisioning a PVC on an Existing Lustre File System with Mount Options
You can optimize the performance and control access to an existing Lustre file system by specifying mount options for the PV. Specifying mount options enables you to fine-tune how pods interact with the file system.
To include mount options:
1.  Start by following the instructions in [Provisioning a PVC on an Existing Lustre File System](#provisioning-a-pvc-on-an-existing-lustre-file-system).
2.  In the PV manifest described in [Provisioning a PVC on an Existing Lustre File System](#provisioning-a-pvc-on-an-existing-lustre-file-system), add the `spec.mountOptions` field, which enables you to specify how the PV should be mounted by pods.
    For example, in the lustre-pv-example.yaml manifest file shown in [Provisioning a PVC on an Existing Lustre File System](#provisioning-a-pvc-on-an-existing-lustre-file-system), you can include the `mountOptions` field as follows:
    ```
    apiVersion: v1
    kind: PersistentVolume
    metadata:
      name: lustre-pv-example
    spec:
      capacity:
        storage: 31.2T
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

### Encrypting Data At Rest on an Existing Lustre File System
The File Storage with Lustre service always encrypts data at rest, using Oracle-managed encryption keys by default. However, you have the option to specify at-rest encryption using your own master encryption keys that you manage yourself in the Vault service.
For more information about File Storage with Lustre file systems that use Oracle-managed encryption keys or your own master encryption keys that you manage yourself, see [Updating File System Encryption](https://docs.oracle.com/iaas/Content/lustre/file-system-encryption.htm).
