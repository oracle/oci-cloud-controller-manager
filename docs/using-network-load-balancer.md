# Network load balancer

## Setup 

1. Make sure you have installed [CCM](../README.md) version v1.19.12 or later

## Create Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: example-nlb
  annotations:
    oci-network-load-balancer.oraclecloud.com/security-list-management-mode: "All"
    oci.oraclecloud.com/load-balancer-type: nlb
spec:
  selector:
    app: example-nlb
  ports:
    - port: 8088
      targetPort: 80
  type: LoadBalancer
  externalTrafficPolicy: Local
```
For more info please refer [OKE NLB DOC][1]

[1]: https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengcreatingloadbalancer.htm#contengcreatingnetworkloadbalancers
