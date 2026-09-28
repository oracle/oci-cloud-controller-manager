# Node filtering

The OCI Cloud Controller Manager can restrict node-scoped reconciliation to
Nodes matching a Kubernetes label selector. This is useful when one Kubernetes
cluster contains Nodes managed by more than one cloud provider.

## Configuration

Set `--node-filter-requirements` on the `oci-cloud-controller-manager`
container. For example:

```yaml
args:
  - --cloud-provider=oci
  - --cloud-config=/etc/oci/cloud-provider.yaml
  - --node-filter-requirements=platform=oci
```

The flag accepts standard Kubernetes label selector syntax, including equality,
inequality, set, existence, and non-existence requirements:

```text
--node-filter-requirements=platform=oci
--node-filter-requirements=platform=oci,!maintenance
--node-filter-requirements=platform in (oci,oke)
```

Multiple comma-separated requirements are combined with logical AND. An empty
selector is the default and preserves the existing behavior of managing all
Nodes. An invalid selector prevents the controller manager from starting.

## Scope

The selector limits the Nodes observed by:

* The Kubernetes cloud node, node lifecycle, service, and route controllers
* The OCI node info, Flex CIDR, and instance tagging controllers
* The OCI Node lister used to construct load balancer backends

You can verify the selected Nodes with the same selector:

```bash
kubectl get nodes -l 'platform=oci'
```
