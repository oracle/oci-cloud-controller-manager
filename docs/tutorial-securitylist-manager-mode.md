# Tutorial

This example will show you how to use the CCM to create a service 
as well as explicitly specifying a security list management mode to
configure how security lists are managed by the CCM.

The possible security list modes:
"All" (default): Manage all required security list rules for load balancer services.
"Frontend":      Manage only security list rules for ingress to the load balancer. 
                 Requires that the user has setup a rule that allows inbound traffic 
                 to the appropriate ports for kube proxy health port, node port ranges, 
                 and health check port ranges.
                 E.g. 10.82.0.0/16 30000-32000.
"None":          Disables all security list management. Requires that the user has setup 
                 a rule that allows inbound traffic to the
                 appropriate ports for kube proxy health port, node port ranges, 
                 and health check port ranges. 
                 E.g. 10.82.0.0/16 30000-32000.
                 Additionally, requires the user to mange rules to allow inbound traffic to load balancers.

Note:
- If an invalid mode is passed in the annotation, then the default ("All") mode is configured.
- If an annotation is not specified, the mode specified in the cloud provider config file is configured.

### Load balancer example

When you create a service with `type: LoadBalancer` an OCI load balancer will
be created.

The example below will create an NGINX deployment, expose it via a load
balancer and disables all security list management. 
Note: 
- The service **type** is set to **LoadBalancer**.
- The annotation must follow the [following format][1] **oci-load-balancer-security-list-management-mode**, 
and declared as `"All"`, `"Frontend"`, `"None"`.

```yaml
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
---
kind: Service
apiVersion: v1
metadata:
  name: nginx-service
  annotations:
    oci-load-balancer-security-list-management-mode: "None"
spec:
  selector:
    app: nginx
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: 80
```
Create it

```bash
$ kubectl create -f examples/nginx-demo-seclist-disabled.yaml
```

Watch the service and await a public IP address. This will be the load balancer
IP which you can use to connect to your service.

```bash
$ kubectl get svc --watch
NAME            CLUSTER-IP     EXTERNAL-IP      PORT(S)        AGE
nginx-service   10.96.97.137   129.213.12.174   80:30274/TCP   5m
```

You can now access your service via the provisioned load balancer

```bash
$ curl -i http://129.213.12.174
```


[1]: https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/load-balancer-annotations.md