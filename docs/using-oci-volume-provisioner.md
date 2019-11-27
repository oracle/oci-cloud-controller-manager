# OCI Volume Provisioner

The OCI Volume Provisioner enables [dynamic provisioning][1] of storage
resources when running Kubernetes on Oracle Cloud Infrastructure (OCI). It uses
the [OCI Flexvolume Driver][2] to bind storage resources to Kubernetes Nodes.
The volume provisioner offers support for [Block Volumes][5] and File System
Storage volumes.


## Deployment

The oci-volume-provisioner is provided as a Kubernetes Deployment.

### Prerequisites

  - Install the [OCI flexvolume driver][2] if provisioning Block Storage
    volumes.


### Configuration

You will first need to create a cloud-provider configuration file tailored to
your OCI account and Kubernetes cluster. Please see the provider configuration
documentation [here][5].

```bash
$ kubectl create secret generic oci-volume-provisioner \
  -n kube-system \
  --from-file=config.yaml=cloud-provider.yaml
```

### OCI Permissions

If authenticating with OCI via an IAM user please ensure it has the following
privileges in the OCI API by creating a [policy][6] tied to a group or user.

```
Allow group <name> to manage volumes in compartment <compartment>
Allow group <name> to manage file-systems in compartment <compartment>
```

## Deploy the OCI Volume Provisioner

First select the release to deploy.

If your cluster is configured to use [RBAC][3] you will need to submit the
following, replacing the <VERSION> placeholder with the selected version:

```
kubectl apply -f https://github.com/oracle/oci-volume-provisioner/releases/download/<VERSION>/oci-volume-provisioner-rbac.yaml
```

Deploy the volume provisioner into your Kubernetes cluster:

```
kubectl apply -f https://github.com/oracle/oci-volume-provisioner/releases/download/<VERSION>/oci-volume-provisioner.yaml
```

Lastly, verify that the oci-volume-provisioner is running in your cluster. By default it runs in the 'kube-system' namespace.

```
kubectl -n kube-system get po | grep volume-provisioner
```

## Usage

In this example we'll use the OCI Volume Provisioner to create persistent
storage for an NGINX Pod.

### Create a StorageClass

```bash
$ cat <<EOF | kubectl apply -f -
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: oci
provisioner: oracle.com/oci
EOF
```

### Create a PersistantVolumeClaim

Next we'll create a [PersistentVolumeClaim][4] (PVC).

The storageClassName must match the "oci" storage class supported by the
provisioner.

The matchLabels should contain the (shortened) Availability Domain (AD) within
which you want to provision the volume. For example in Phoenix that might be
`PHX-AD-1`, in Ashburn `US-ASHBURN-AD-1`, in Frankfurt `EU-FRANKFURT-1-AD-1`,
and in London `UK-LONDON-1-AD-1`.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: my-volume
spec:
  storageClassName: "oci"
  selector:
    matchLabels:
      failure-domain.beta.kubernetes.io/zone: "PHX-AD-1"
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```

After submitting the PVC, you should see a block storage volume available in
your OCI tenancy.

### Create a Kubernetes Pod that references the PVC

Now you have a PVC, you can create a Kubernetes Pod that will consume the
storage.

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: nginx
spec:
  volumes:
    - name: nginx-storage
      persistentVolumeClaim:
        claimName: my-volume
  containers:
    - name: nginx
      image: nginx
      ports:
        - containerPort: 80
      volumeMounts:
      - mountPath: /usr/share/nginx/html
        name: nginx-storage
```

### Create a block volume from a backup

You can use annotations to create a volume from an existing backup. Simply use
an annotation and reference the volume OCID.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: block-volume-from-backup
  annotations:
    volume.beta.kubernetes.io/oci-volume-source: ocid...
spec:
  storageClassName: "oci"
  selector:
    matchLabels:
      failure-domain.beta.kubernetes.io/zone: PHX-AD-1
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```

## Misc

You can add a prefix to volume display names by setting an
`OCI_VOLUME_NAME_PREFIX` environment variable.

[1]: http://blog.kubernetes.io/2016/10/dynamic-provisioning-and-storage-in-kubernetes.html
[2]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/docs/using-oci-flexvolume-driver.md
[3]: https://kubernetes.io/docs/admin/authorization/rbac/
[4]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims
[5]: https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm
[6]: https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policysyntax.htm
