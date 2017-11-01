# Load Balancer Annotations

This file defines a list of [Service][4] `type: LoadBalancer` annotations which are
supported by the `oci-cloud-controller-manager`.

All annotations are prefixed with `service.beta.kubernetes.io/`.

They are added to the Service metadata as follows:

```yaml
kind: Service
apiVersion: v1
metadata:
  name: nginx-service
  annotations:
    service.beta.kubernetes.io/oci-load-balancer-internal: "443"
spec:
  ...
```

## Load balancer properties

| Name | Description | Default |
| `oci-load-balancer-internal` | Create an [internal load balancer][1]. Cannot be modified after load balancer creation. | `false` |
| `oci-load-balancer-shape` | A template that determines the load balancer's total pre-provisioned maximum capacity (bandwidth) for ingress plus egress traffic. Available shapes include `100Mbps`, `400Mbps`, and `8000Mbps.` Cannot be modified after load balancer creation. | `"100Mbps"` |
| `oci-load-balancer-subnet1` | The OCID of the first [subnet][2] of the two required subnets to attach the load balancer to. Must be in separate Availability Domains. | Value provided in config file |
| `oci-load-balancer-subnet2` | The OCID of the second [subnet][2] of the two required subnets to attach the load balancer to. Must be in separate Availability Domains. | Value provided in config file |

## TLS-related

| Name | Description | Default |
| ---- | ----------- | ------- |
| `oci-load-balancer-tls-secret` | A reference in the form `<namespace>/<secretName>` to a Kubernetes [TLS secret][3]. | `""` |
| `oci-load-balancer-ssl-ports` | A `,` separated list of port number(s) for which to enable SSL termination. | `""` |

[1]: https://kubernetes.io/docs/concepts/services-networking/service/#internal-load-balancer
[2]: https://docs.us-phoenix-1.oraclecloud.com/Content/Network/Tasks/managingVCNs.htm
[3]: https://kubernetes.io/docs/concepts/services-networking/ingress/#tls
[4]: https://kubernetes.io/docs/concepts/services-networking/service/
