# OCI Flexvolume Driver

The Oracle Cloud Infrastructure (OCI) Flexvolume driver enables attaching of
[OCI Block Storage][1] volumes to Kubernetes Nodes and mounting these volumes
into Pods via the [Flexvolume][2] plugin interface.

We recommend you use this driver in conjunction with the OCI Volume Provisioner.
See the [oci-volume-provisioner][3] for more information.

## Deployment

#### Configuration

You will first need to create a cloud-provider configuration file tailored to
your OCI account and Kubernetes cluster. Please see the provider configuration
documentation [here][5].

```bash
$ kubectl create secret generic oci-volume-provisioner \
  -n kube-system \
  --from-file=config.yaml=cloud-provider.yaml
```

### Kubernetes API Access

The flexvolume driver requires access to the Kubernetes API to resolve Node
names to the corosponding Instance OCID. A kubeconfig file with appropriate
permissions (RBAC: nodes/get) needs to be provided via a Kubernetes Secret

```
kubectl create secret generic oci-flexvolume-driver-kubeconfig \
    -n kube-system \
    --from-file=kubeconfig=kubeconfig
```

### Using instance principals

To authenticate using [instance principals][8] the following policies must first be
applied to the dynamic group of instances that intend to use the flexvolume driver:

TODO: The below is now out of date as we required mounting in a kubeconfig.

```
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read vnic-attachments in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read vnics in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read instances in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to read subnets in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to use volumes in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to use instances in compartment id ${var.compartment_ocid}",
"Allow group id ${oci_identity_group.flexvolume_driver_group.id} to manage volume-attachments in compartment id ${var.compartment_ocid}",
```

### Environment Variables

You can set these in the environment to override the default values.

| VAR | Default | Description |
| --- | ------- | ----------- |
| `OCI_FLEXD_DRIVER_LOG_DIR`  | `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci` | Directory where the log file is written. |
| `OCI_FLEXD_DRIVER_DIRECTORY` | `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci` | Directory where the driver binary lives |
| `OCI_FLEXD_CONFIG_DIRECTORY` | `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci` | Directory where the driver configuration lives |
| `OCI_FLEXD_KUBECONFIG_PATH` | `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/kubeconfig` | An override to allow the fully qualified path of the kubeconfig resource file to be specified. This take precedence over additional configuration. |

## OCI Policies

You must ensure the user (or group) associated with the OCI credentials provided has the following level of access. See [policies][7] for more information.

TODO: The below is now out of date as we require mounting in a kubeconfig.

```
"Allow group id GROUP to read vnic-attachments in compartment id COMPARTMENT",
"Allow group id GROUP to read vnics in compartment id COMPARTMENT"
"Allow group id GROUP to read instances in compartment id COMPARTMENT"
"Allow group id GROUP to read subnets in compartment id COMPARTMENT"
"Allow group id GROUP to use volumes in compartment id COMPARTMENT"
"Allow group id GROUP to use instances in compartment id COMPARTMENT"
"Allow group id GROUP to manage volume-attachments in compartment id COMPARTMENT"
```

### Kubernetes DaemonSet Installer

The recommended way to install the driver is through the DaemonSet installer
mechanism. This will create two DaemonSets, one specifically for master nodes,
allowing configuration via a Kubernetes Secret, and one for worker nodes.

```bash
$ kubectl apply -f https://github.com/oracle/oci-flexvolume-driver/releases/download/${flexvolume_driver_version}/rbac.yaml
$ kubectl apply -f https://github.com/oracle/oci-flexvolume-driver/releases/download/${flexvolume_driver_version}/oci-flexvolume-driver.yaml
```

You'll still need to add the config file as a Kubernetes Secret.

Once the Secret is set and the daemonsets deployed, the configuration file will be placed onto the master nodes.

## Assumptions

- If a Flexvolume is specified for a Pod, it will only work with a single
  replica. (or if there is more than one replica for a Pod, they will all have
  to run on the same Kubernetes Node). This is because a volume can only be
  attached to one instance at any one time. Note: This is in common with both
  the Amazon and Google persistent volume implementations, which also have the
  same constraint.

- If nodes in the cluster span availability domain you must make sure your Pods are scheduled
  in the correct availability domain. This can be achieved using the label selectors with the zone/region.

  Using the [oci-volume-provisioner][3] makes this much easier.

- For all nodes in the cluster, the instance display name in the OCI API must
  match with the instance hostname, start with the vnic hostnamelabel or match the public IP.
  This relies on the requirement that the nodename must be resolvable.

## Limitations

Due to [kubernetes/kubernetes#44737][6] ("Flex volumes which implement
`getvolumename` API are getting unmounted during run time") we cannot implement
`getvolumename`. From the issue:

> Detach call uses volume name, so the plugin detach has to work with PV Name

This means that the Persistent Volume (PV) name in the `pod.yml` _must_ be
the last part of the block volume OCID ('.' separated). Otherwise, we would
have no way of determining which volume to detach from which worker node. Even
if we were to store state at the time of volume attachment PV names would have
to be unique across the cluster which is an unreasonable constraint.

The full OCID cannot be used because the PV name must be shorter than 63
characters and cannot contain '.'s. To reconstruct the OCID we use the region
of the master on which `Detach()` is exected so this blocks support for cross
region clusters.

[1]: https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm
[2]: https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md
[3]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/docs/using-oci-volume-provisioner.md
[4]: https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Tasks/gettingmetadata.htm
[6]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/docs/using-oci-volume-provisioner.md
[5]: https://docs.us-phoenix-1.oraclecloud.com/Content/API/SDKDocs/cli.htm
[6]: https://github.com/kubernetes/kubernetes/issues/44737
[7]: https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policies.htm
[8]: https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Tasks/callingservicesfrominstances.htm
