# Install

A guide to installing the kubernetes-cloud-controller-manager. To get the CCM
running in your Kubernetes cluster you will need to do the following:

 1. Prepare your Kubernetes cluster for running an external cloud provider.
 2. Configure the authentication information (secrets etc.) needed to access the
    BMCS API.
 3. Deploy the CCM as a [daemon set][1]

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

## Cloud Controller Manager

The following steps will configure the CCM to run in your cluster

### Create namespace

```
$ kubectl create namespace bmcs
```

#### Create authentication and configuration

```
$ kubectl -n bmcs create secret generic bmcs-api-key --from-file="$HOME/.oraclebmc/bmcs_api_key.pem"

# We publish the kubernetes-cloud-controller-mananger to a private Docker registry.
# You will need a secret to pull images from it.
$ kubectl -n bmcs create secret docker-registry odx-docker-pull-secret \
    --docker-server="registry.oracledx.com" \
    --docker-username="agent" \
    --docker-password="$REGISTRY_PASSWORD" \
    --docker-email="k8s@oracle.com"

$ kubectl create -f manifests/kubernetes-cloud-controller-manager-config.yaml
```

### Deploy the Kubernetes Cloud Controller Manager

We publish tagged builds of the CCM in the ODX registry. The cloud controller manager runs as a daemonset and uses leader election to
ensure that only one instance of the controller is running at a given time.

```
$ kubectl apply -f manifests/kubernetes-cloud-controller-manager.yaml
```

Check the CCM logs to ensure it's running correctly

```
$ kubectl get po -n bmcs
NAME                                     READY     STATUS    RESTARTS   AGE
bmcs-cloud-controller-manager-ds-k2txq   1/1       Running   0          19s

$ kubectl logs bmcs-cloud-controller-manager-ds-k2txq -n bmcs
I0905 13:44:51.785964       7 flags.go:52] FLAG: --address="0.0.0.0"
I0905 13:44:51.786063       7 flags.go:52] FLAG: --allocate-node-cidrs="false"
I0905 13:44:51.786074       7 flags.go:52] FLAG: --alsologtostderr="false"
I0905 13:44:51.786078       7 flags.go:52] FLAG: --cloud-config="/etc/bmcs/cloud-config.cfg"
I0905 13:44:51.786083       7 flags.go:52] FLAG: --cloud-provider="external"
```

# Uninstall

To uninstall everything simply delete the bmcs namespace

```
kubectl delete ns bmcs
```

[1]: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
[2]: https://github.com/kubernetes/kubernetes/pull/49017
