# Flex CIDR Provider Controller

The OCI Cloud Controller Manager (OCI CCM) Flex CIDR Provider Controller enables self-managed Kubernetes clusters running on Oracle Cloud Infrastructure (OCI) to allocate and advertise a VCN-native IP CIDR block that pods can use on a node. This enables using third part CNIs like Cilium to be used in native (direct routing) mode.

The Flex CIDR Provider Controller dynamically allocates per-node Pod CIDRs based on OCI Compute instance metadata and publishes them to the Kubernetes Node object. The CNI solution e.g. Cilium consumes these Pod CIDRs using Kubernetes Host-Scope IPAM, allowing pods to communicate directly over the OCI Virtual Cloud Network (VCN) without overlay encapsulation.

## Architecture

```text
                OCI Instance Metadata
                        │
                        ▼
        OCI Cloud Controller Manager
          (Flex CIDR Provider)
                        │
              Calculates Pod CIDRs
                        │
                        ▼
      Kubernetes Node.spec.podCIDRs
                        │
                        ▼
      CNI Kubernetes Host-Scope IPAM e.g. Cilium
                        │
                        ▼
             Pods receive VCN-native IPs
```

## Prerequisites

- Self-managed (non-OKE) Kubernetes cluster running on OCI
- Kubernetes v1.33 or later
- containerd or CRI-O runtime
- Kubernetes Host-Scope compatabile CNI which used podCIDR installed as the cluster CNI e.g. Cilium
- OCI CLI configured
- Permission to modify OCI Compute instance metadata
- OCI IAM Dynamic Groups and Policies configured for worker nodes

## Deploy OCI Cloud Controller Manager with the Flex CIDR Provider Controller enabled

Deploy OCI Cloud Controller Manager using the standard installation instructions. The OCI CCM configuration file, i.e. oci-cloud-configuration-manager.yaml has the following new configuration for the FlexCIDR provider controller. By default, it is set to false. Change it to true.

```yaml
env:
- name: ENABLE_FLEX_CIDR_CONTROLLER
  value: "true"
```

## Configure IAM Permissions

OCI recommends using Instance Principal based access for the CCM. The policy requires granting the ability to use virtual-network-family in the worker node compartment. To use the FlexCIDR Provider Controller, it should grant the ability to manage virtual-network-family in the worker node compartment. 

Example policy:

```text
Allow dynamic-group <worker-node-dynamic-group> to manage virtual-network-family in compartment <worker-node-compartment>
```

## Configure Worker Node Metadata

The OKE FlexCIDR Provider will dynamically size and allocate Pod CIDR blocks for Kubernetes nodes based on user-specified IP requirements. It enables flexible IP planning by allowing users to define: 

- ip-count: number of IPs needed for pods on a node. 
- cidr-blocks: [optional] user-specified parent CIDRs to carve pod CIDRs from. 

The OKE FlexCIDR Provider expects these values to be stored as metadata on each worker node instance. It fetches this metadata via the OCI Instance Metadata Service (IMDS) and calculates the correct IPv4 CIDR blocks (and IPv6 prefixes) to allocate for that node’s primary VNIC. Use the OCI CLI to configure this metadata: 

```bash
oci compute instance update \
  --instance-id <instance_ocid> \
  --metadata '{
    "flexcidr-primary-vnic": {
      "ip-count": 32,
      "cidr-blocks": [
        "10.0.10.0/24"
      ]
    }
  }'
```

## Pod CIDR Allocation

```
prefix = 32 - log2(ip-count)
```

| Requested Pod IPs | Assigned CIDR |
|-------------------|---------------|
| 32 | /27 |
| 64 | /26 |
| 128 | /25 |
| 256 | /24 |

OCI CCM patches `Node.spec.podCIDRs` with the calculated CIDRs.

## Configure CNI
Below are abbrevieated examples of configuring the Cilium CNI in different native routing setups when using the OCI FlexCIDR Provider Controller 

VCN-native pod routing CIDR i.e. the super-net from which OKE FlexCIDR Provider Controller carves pod CIDR blocks

```
10.0.10.0/16
```

### Native Routing with kube-proxy

```yaml
cluster:
  name: kubernetes

operator:
  replicas: 1

routingMode: native
ipv4NativeRoutingCIDR: 10.0.10.0/16

ipam:
  mode: kubernetes

ipv4:
  enabled: true

ipv6:
  enabled: false

kubeProxyReplacement: "false"
```

### Native Routing with eBPF kube-proxy replacement

```yaml
cluster:
  name: kubernetes

operator:
  replicas: 1

routingMode: native
ipv4NativeRoutingCIDR: 10.0.10.0/16

ipam:
  mode: kubernetes

ipv4:
  enabled: true

ipv6:
  enabled: false

kubeProxyReplacement: true

k8sServiceHost: <API_SERVER_IP>
k8sServicePort: 6443
```

The above examples contain only the relevant subset of the full configuration required for Cilium. You should refer to the Cilium [documentation](https://docs.cilium.io/en/stable/) for complete details on configuring [Cilium with Kubernetes Host Scope IPAM](https://docs.cilium.io/en/stable/network/concepts/ipam/kubernetes/). 

## Verification

```bash
kubectl -n kube-system get pods
```

```bash
kubectl get node <node-name> -o jsonpath='{.spec.podCIDRs}{"\n"}'
```

```bash
cilium status
```

Verify that:

- OCI CCM is running.
- Node `spec.podCIDRs` are populated.
- Pods receive addresses from the assigned Pod CIDRs.
- Cilium is operating in native routing mode.

## Troubleshooting

### Pod CIDRs are not assigned

Verify:

- `flexcidr-primary-vnic` metadata exists.
- `ip-count` is valid.
- OCI CCM is running.
- IAM policies allow `manage virtual-network-family`.

Inspect logs:

```bash
kubectl logs -n kube-system <oci-cloud-controller-manager-pod>
```

### Cilium cannot reach kubelet

Ensure TCP/10250 is allowed by Security Lists, NSGs, and local firewall rules.

### Network Requirements

- Node-to-node traffic for Pod CIDRs.
- TCP/6443 from workers to the Kubernetes API server.
- TCP/10250 when kubelet access is required.
- Route tables include the Flex CIDR ranges when traffic exits the VCN.
