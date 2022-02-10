# OCI Flexvolume Driver

This project implements a flexvolume driver for Kubernetes clusters
running on Oracle Cloud Infrastructure (OCI). It enables mounting of [OCI block
storage volumes][1] to Kubernetes Pods via the [Flexvolume][2] plugin interface.

We recommend you use this driver in conjunction with the OCI Volume Provisioner.
See the [oci-volume-provisioner](/flex-volume-provisioner.md) for more information.

## Install / Setup

We publish the OCI flexvolume driver as a single binary that needs to be
installed on every node in your Kubernetes cluster.

### Kubernetes DaemonSet Installer

The recommended way to install the driver is through the DaemonSet installer mechanism. This will create two daemonsets, one specifically for master nodes, allowing configuration via a Kubernetes Secret, and one for worker nodes.

```bash
$ export RELEASE=?
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-flexvolume-driver.yaml
$ kubectl apply -f https://github.com/oracle/oci-cloud-controller-manager/releases/download/${RELEASE}/oci-flexvolume-driver-rbac.yaml
```

You'll still need to add the config file as a Kubernetes Secret.

#### Configuration

The driver requires API credentials for a OCI account with the ability
to attach and detach [OCI block storage volumes][1] from to/from the appropriate
nodes in the cluster.

These credentials should be provided via a YAML file present on **master** nodes
in the cluster at `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/config.yaml`
in the format as in the example file [here](/manifests/provider-config-example.yaml)

If `"region"` and/or `"compartment"` are not specified in the config file
they will be retrieved from the hosts [OCI metadata service][4].

### Submit configuration as a Kubernetes secret

The configuration file above can be submitted as a Kubernetes Secret onto the master nodes.

```
kubectl create secret generic oci-flexvolume-driver \
    -n kube-system \
    --from-file=config.yaml=config.yaml
```

Once the Secret is set and the daemonsets deployed, the configuration file will be placed onto the master nodes.

##### Using instance principals

To authenticate using [instance principals][9] the following policies must first be 
applied to the dynamic group of instances that intend to use the flexvolume driver:

```
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read vnic-attachments in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read vnics in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read instances in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read subnets in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to use volumes in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to use instances in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to manage volume-attachments in compartment id ${var.compartment_ocid}",
```

The configuration file requires a simple configuration in the following format:

```yaml
---
useInstancePrincipals: true
```

#### Driver Kubernetes API Access

The driver needs to get node information from the Kubernetes API server. A kubeconfig file with appropriate permissions (rbac: nodes/get) needs
to be provided in the same manor as the OCI auth config file above.

```
kubectl create secret generic oci-flexvolume-driver-kubeconfig \
    -n kube-system \
    --from-file=kubeconfig=kubeconfig
```

Once the Secret is set and the DaemonSet deployed, the kubeconfig file will be placed onto the master nodes.

#### Extra configuration values

You can set these in the environment to override the default values.

* `OCI_FLEXD_DRIVER_LOG_DIR` - Directory where the log file is written (Default: `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci`)
* `OCI_FLEXD_DRIVER_DIRECTORY` - Directory where the driver binary lives (Default:
`/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci`)
* `OCI_FLEXD_CONFIG_DIRECTORY` - Directory where the driver configuration lives (Default:
`/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci`)
* `OCI_FLEXD_KUBECONFIG_PATH` - An override to allow the fully qualified path of the kubeconfig resource file to be specified. This take precedence over additional configuration.

## OCI Policies

You must ensure the user (or group) associated with the OCI credentials provided has the following level of access. See [policies][8] for more information.

```
"Allow group id GROUP to read vnic-attachments in compartment id COMPARTMENT",
"Allow group id GROUP to read vnics in compartment id COMPARTMENT"
"Allow group id GROUP to read instances in compartment id COMPARTMENT"
"Allow group id GROUP to read subnets in compartment id COMPARTMENT"
"Allow group id GROUP to use volumes in compartment id COMPARTMENT"
"Allow group id GROUP to use instances in compartment id COMPARTMENT"
"Allow group id GROUP to manage volume-attachments in compartment id COMPARTMENT"
```

## Tutorial

This guide will walk you through creating a Pod with persistent storage. It assumes
that you have already installed the flexvolume driver in your cluster.

1. Create a block storage volume. This can be done using the `oci` [CLI][5] as follows:

```bash
$ oci bv volume create \
    --availability-domain="aaaa:PHX-AD-1" \
    --compartment-id "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
```

2. Add a volume to your `pod.yml` in the format below and named with the last
   section of your volume's OCID (see limitations). E.g. a volume with the OCID

```
ocid1.volume.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

Would be named `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`
in the pod.yml as shown below.

```yaml
volumes:
  - name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    flexVolume:
      driver: "oracle/oci"
      fsType: "ext4"
```

3. Add volume mount(s) in the appropriate container(s) in your as follows:

```yaml
volumeMounts:
  - name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    mountPath: /usr/share/nginx/html
```
(Where `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"` is the
last '.' sparated section of the volume OCID.)

### Fixing a Pod to a Node

It's important to note that a block volume can only be attached to a Node that runs
in the same AD. To get around this problem, you can use a nodeSelector to ensure
that a Pod is scheduled on a particular Node.

This following example shows you how to do this.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx
    ports:
    - containerPort: 80
    volumeMounts:
    - name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
      mountPath: /usr/share/nginx/html
  nodeSelector:
    node.info/availability.domain: 'UpwH-US-ASHBURN-AD-1'
  volumes:
  - name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    flexVolume:
      driver: "oracle/oci"
      fsType: "ext4"
```

## Debugging

The flexvolume driver writes logs to `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/oci_flexvolume_driver.log` by default.

## Assumptions

- If a Flexvolume is specified for a Pod, it will only work with a single
  replica. (or if there is more than one replica for a Pod, they will all have
  to run on the same Kubernetes Node). This is because a volume can only be
  attached to one instance at any one time. Note: This is in common with both
  the Amazon and Google persistent volume implementations, which also have the
  same constraint.

- If nodes in the cluster span availability domain you must make sure your Pods are scheduled
  in the correct availability domain. This can be achieved using the label selectors with the zone/region.

  Using the [oci-volume-provisioner](/flex-volume-provisioner.md)
  makes this much easier.

- For all nodes in the cluster, the instance display name in the OCI API must
  match with the instance hostname, start with the vnic hostnamelabel or match the public IP.
  This relies on the requirement that the nodename must be resolvable.


[1]: https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm
[2]: https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md
[3]: https://github.com/oracle/oci-volume-provisioner/
[4]: https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Tasks/gettingmetadata.htm
[5]: https://docs.us-phoenix-1.oraclecloud.com/Content/API/SDKDocs/cli.htm
[6]: https://github.com/oracle/oci-volume-provisioner
[7]: https://github.com/kubernetes/kubernetes/issues/44737
[8]: https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policies.htm
[9]: https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Tasks/callingservicesfrominstances.htm
