# Load Balancer Annotations

This file defines a list of [Service][4] `type: LoadBalancer` annotations which are
supported by the `oci-cloud-controller-manager`.

All annotations are prefixed with `service.beta.kubernetes.io/`. For example:

```yaml
kind: Service
apiVersion: v1
metadata:
  name: nginx-service
  annotations:
    service.beta.kubernetes.io/oci-load-balancer-shape: "400Mbps"
    service.beta.kubernetes.io/oci-load-balancer-subnet1: "ocid..."
    service.beta.kubernetes.io/oci-load-balancer-subnet2: "ocid..."
spec:
  ...
```

## Load balancer properties

| Name                                        | Description                                                                                                                                                                                                                                        | Default                                          |
| -----                                       | -----------                                                                                                                                                                                                                                        | -------                                          |
| `oci-load-balancer-internal`                | Create an [internal load balancer][1]. Cannot be modified after load balancer creation.                                                                                                                                                            | `false`                                          |
| `oci-load-balancer-shape`                   | A template that determines the load balancer's total pre-provisioned maximum capacity (bandwidth) for ingress plus egress traffic. Available shapes include `100Mbps`, `400Mbps`, and `8000Mbps.` Cannot be modified after load balancer creation. | `"100Mbps"`                                      |
| `oci-load-balancer-subnet1`                 | The OCID of the first [subnet][2] of the two required subnets to attach the load balancer to. Must be in separate Availability Domains.                                                                                                            | Value provided in config file                    |
| `oci-load-balancer-subnet2`                 | The OCID of the second [subnet][2] of the two required subnets to attach the load balancer to. Must be in separate Availability Domains.                                                                                                           | Value provided in config file                    |
| `oci-load-balancer-connection-idle-timeout` | The maximum idle time, in seconds, allowed between two successive receive or two successive send operations between the client and backend servers.                                                                                                | `300` for TCP listeners, `60` for HTTP listeners |
| `oci-load-balancer-security-list-management-mode` | Specifies the [security list mode](##security-list-management-modes) (`"All"`, `"Frontend"`,`"None"`) to configure how security lists are managed by the CCM.                            | `"All"`            
| `oci-load-balancer-backend-protocol` | Specify protocol on which the listener accepts connection requests. To get a list of valid protocols, use the [`ListProtocols`][5] operation.                          | `"TCP"`            

## TLS-related

| Name | Description | Default |
| ---- | ----------- | ------- |
| `oci-load-balancer-tls-secret` | A reference in the form `<namespace>/<secretName>` to a Kubernetes [TLS secret][3]. | `""` |
| `oci-load-balancer-ssl-ports` | A `,` separated list of port number(s) for which to enable SSL termination. | `""` |

## Security List Management Modes
| Mode | Description | 
| ---- | ----------- | 
| `"All"` | CCM will manage all required security list rules for load balancer services | 
| `"Frontend"` | CCM will manage  only security list rules for ingress to the load balancer. Requires that the user has setup a rule that allows inbound traffic to the appropriate ports for kube proxy health port, node port ranges, and health check port ranges.  | 
| `"None`" | Disables all security list management. Requires that the user has setup a rule that allows inbound traffic to the appropriate ports for kube proxy health port, node port ranges, and health check port ranges. *Additionally, requires the user to mange rules to allow inbound traffic to load balancers.* | 

Note:
- If an invalid mode is passed in the annotation, then the default (`"All"`) mode is configured.
- If an annotation is not specified, the mode specified in the cloud provider config file is configured.  

[1]: https://kubernetes.io/docs/concepts/services-networking/service/#internal-load-balancer
[2]: https://docs.us-phoenix-1.oraclecloud.com/Content/Network/Tasks/managingVCNs.htm
[3]: https://kubernetes.io/docs/concepts/services-networking/ingress/#tls
[4]: https://kubernetes.io/docs/concepts/services-networking/service/
[5]: https://docs.cloud.oracle.com/iaas/api/#/en/loadbalancer/20170115/LoadBalancerProtocol/ListProtocols
