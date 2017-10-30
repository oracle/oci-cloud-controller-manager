# Install

A guide to installing the `oci-cloud-controller-manager`.

To get the CCM running in your Kubernetes cluster you will need to do the
following:

 1. Prepare your Kubernetes cluster for running an external cloud provider.
 2. Create a Kubernetes secret containing the configuration for the CCM.
 3. Deploy the CCM as a [DaemonSet][1].

## Preparing Your Cluster

To deploy the Cloud Controller Manager (CCM) your cluster must be configured to
use an external cloud-provider.

This involves:
 - Setting the `--cloud-provider=external` flag on the `kubelet` on **all
   nodes** in your cluster.
 - Setting the `--cloud-provider=external` flag on the `kube-controller-manager`
   in your Kubernetes control plane.

**Depending on how kube-proxy is run you "may" need the following:**

- Ensuring that `kube-proxy` tolerates the uninitialised cloud taint. The
  following should appear in the `kube-proxy` pod yaml:

```yaml
- effect: NoSchedule
  key: node.cloudprovider.kubernetes.io/uninitialized
  value: "true"
```

If your cluster was created using `kubeadm` >= v1.7.2 this toleration will
already be applied. See [kubernetes/kubernetes#49017][2] for details.

Remember to restart any components that you have reconfigured before continuing.

## Authentication and Configuration

We publish the `oci-cloud-controller-mananger` to a private Docker registry. You
will need a [Docker registry secret][4] to pull images from it.

```bash
$ kubectl -n kube-system create secret docker-registry wcr-docker-pull-secret \
    --docker-server="wcr.io" \
    --docker-username="$DOCKER_REGISTRY_USERNAME" \
    --docker-password="$DOCKER_REGISTRY_PASSWORD" \
    --docker-email="k8s@oracle.com"
```

An example configuration file can be found [here][3]. Download this file and
populate it with values specific to your chosen OCI identity and tenancy.
Then create the Kubernetes secret with the following command:

```bash
$ kubectl  create secret generic oci-cloud-controller-manager \
     -n kube-system                                           \
     --from-file=cloud-provider.yaml=cloud-provider-example.yaml
```

Note that you must ensure the secret contains the key `cloud-provider.yaml`
rather than the name of the file on disk.

## Deployment

Lastly deploy the controller manager and associated RBAC rules if your cluster
is configured to use RBAC:

```bash
$ kubectl apply -f https://raw.githubusercontent.com/oracle/oci-cloud-controller-manager/master/manifests/oci-cloud-controller-manager.yaml
$ kubectl apply -f https://raw.githubusercontent.com/oracle/oci-cloud-controller-manager/master/manifests/oci-cloud-controller-manager-rbac.yaml
```

Check the CCM logs to ensure it's running correctly:

```
$ kubectl -n kube-system get po | grep oci
oci-cloud-controller-manager-ds-k2txq   1/1       Running   0          19s

$ kubectl -n kube-system logs oci-cloud-controller-manager-ds-k2txq
I0905 13:44:51.785964       7 flags.go:52] FLAG: --address="0.0.0.0"
I0905 13:44:51.786063       7 flags.go:52] FLAG: --allocate-node-cidrs="false"
I0905 13:44:51.786074       7 flags.go:52] FLAG: --alsologtostderr="false"
I0905 13:44:51.786078       7 flags.go:52] FLAG: --cloud-config="/etc/oci/cloud-config.cfg"
I0905 13:44:51.786083       7 flags.go:52] FLAG: --cloud-provider="oci"
```

# Uninstall

TODO

[1]: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
[2]: https://github.com/kubernetes/kubernetes/pull/49017
[3]: https://github.com/oracle/oci-cloud-controller-manager/tree/master/manifests/cloud-provider-example.yaml
[4]: https://kubernetes.io/docs/concepts/containers/images/#creating-a-secret-with-a-docker-config
