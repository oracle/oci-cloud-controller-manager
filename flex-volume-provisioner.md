# OCI Volume Provisioner

The OCI Volume Provisioner enables [dynamic provisioning][1] of storage resources when running Kubernetes on Oracle Cloud Infrastructure.
It uses the [OCI Flexvolume Driver](/flex-volume-driver.md) to bind storage resources to Kubernetes nodes. The volume provisioner offers support for

* [Block Volumes][5]

## Install

The oci-volume-provisioner is provided as a Kubernetes deployment.

### Submit configuration as a Kubernetes secret

Create a config.yaml file with contents similar to the following. This file will contain authentication
information necessary to authenticate with the OCI APIs and provision block storage volumes.
An example configuration file can be found [here](manifests/provider-config-example.yaml)

Submit this as a Kubernetes Secret.

```bash
kubectl create secret generic oci-volume-provisioner \
    -n kube-system \
    --from-file=config.yaml=config.yaml
```

### OCI Permissions

Please ensure that the credentials used in the secret have the following privileges in the OCI API by creating a [policy](https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policysyntax.htm) tied to a group or user.

```
Allow group <name> to manage volumes in compartment <compartment>
Allow group <name> to manage file-systems in compartment <compartment>
```


## Deploy the OCI Volume Provisioner

Deploy the volume provisioner and associated RBAC rules if your cluster is configured to use [RBAC][3]

```bash
$ export RELEASE=?
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-volume-provisioner.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-volume-provisioner-rbac.yaml
```

Deploy the volume provisioner storage classes:

```
$ kubectl apply -f https://raw.githubusercontent.com/oracle/oci-cloud-controller-manager/master/manifests/volume-provisioner/storage-class.yaml
```

Lastly, verify that the oci-volume-provisioner is running in your cluster. By default it runs in the 'kube-system' namespace.

```
$ kubectl -n kube-system get po | grep oci-block-volume-provisioner
```

## Tutorial

In this example we'll use the OCI Volume Provisioner to create persistent storage for an NGINX Pod.

### Create a PVC

Next we'll create a [PersistentVolumeClaim][4] (PVC).

The storageClassName must match the "oci" storage class supported by the provisioner.

The matchLabels should contain the (shortened) Availability Domain (AD) within
which you want to provision the volume. For example in Phoenix that might be
`PHX-AD-1`, in Ashburn `US-ASHBURN-AD-1`, in Frankfurt `EU-FRANKFURT-1-AD-1`,
and in London `UK-LONDON-1-AD-1`.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: nginx-volume
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

After submitting the PVC, you should see a block storage volume available in your OCI tenancy.

### Create a Kubernetes Pod that references the PVC

Now you have a PVC, you can create a Kubernetes Pod that will consume the storage.

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: nginx
spec:
  volumes:
    - name: nginx-storage
      persistentVolumeClaim:
        claimName: nginx-volume
  containers:
    - name: nginx
      image: nginx
      ports:
        - containerPort: 80
      volumeMounts:
      - mountPath: "/usr/share/nginx/html"
        name: nginx-storage
```

### Create a block volume from a backup

You can use annotations to create a volume from an existing backup. Simply use an annotation and reference the volume OCID.

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
      failure-domain.beta.kubernetes.io/zone: "PHX-AD-1"
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
```

## Misc

You can add a prefix to volume display names by setting an `OCI_VOLUME_NAME_PREFIX` environment variable.


[1]: http://blog.kubernetes.io/2016/10/dynamic-provisioning-and-storage-in-kubernetes.html
[2]: https://github.com/oracle/oci-flexvolume-driver
[3]: https://kubernetes.io/docs/admin/authorization/rbac/
[4]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims
[5]: https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm
