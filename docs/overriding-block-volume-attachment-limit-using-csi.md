# Override the Block Volume Attachment Limit Using CSI

The block volume CSI node driver reports the maximum number of block volumes that Kubernetes can attach to each node. By default, the published CSI manifests set the `--volume-attachment-limit` node-driver flag to `32`.

If nodes in a cluster support different attachment limits, set the `oci.oraclecloud.com/volume-attachment-limit-override` label on an individual node. A valid label value takes precedence over `--volume-attachment-limit` for that node.

## Configure the cluster default

For a Helm deployment, set `volumeAttachmentLimit` to configure the default for nodes without an override label:

```bash
helm upgrade <release-name> manifests/container-storage-interface/csi \
  --set volumeAttachmentLimit=64
```

For the published YAML deployment, update the `--volume-attachment-limit` argument in the `oci-csi-node-driver` DaemonSet and apply the manifest. The value must be a positive integer.

## Set an override

Set the label value to a positive base-10 integer. For example, to advertise a limit of `64` on a node:

```bash
kubectl label node <node-name> \
  oci.oraclecloud.com/volume-attachment-limit-override=64 --overwrite
```

The CSI node driver caches node metadata after its first lookup. After adding, changing, or removing the label, restart the CSI node pod on the affected node so that the CSI node registration is refreshed:

```bash
kubectl -n kube-system get pods -l app=csi-oci-node \
  --field-selector=spec.nodeName=<node-name>
kubectl -n kube-system delete pod <csi-node-pod-name>
```

For a Helm deployment that uses `customHandle`, use the corresponding CSI node pod label and pod name. The DaemonSet recreates the deleted pod automatically.

Verify the advertised limit after the replacement pod is ready:

```bash
kubectl get csinode <node-name> -o yaml
```

The `blockvolume.csi.oraclecloud.com` driver entry should contain the configured value under `allocatable.count`.

## Validation and scope

Only positive integers are accepted. If the label is missing, zero, negative, or non-numeric, the CSI node driver logs a warning for an invalid value and uses `--volume-attachment-limit` instead.

This label applies only to the `blockvolume.csi.oraclecloud.com` driver. It changes the attachment limit advertised to Kubernetes; it does not increase the number of block volumes supported by the underlying OCI instance. Set the value no higher than the attachment capacity of the node's instance shape.

## Remove an override

Remove the label and restart the CSI node pod on that node to return to the configured default:

```bash
kubectl label node <node-name> \
  oci.oraclecloud.com/volume-attachment-limit-override-
```
