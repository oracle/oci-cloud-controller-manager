# Tutorial

This example will show you how to use the CCM to create a load balancer with SSL
termination.

### Load balancer with SSL termination example

When you create a service with --type=LoadBalancer a OCI load balancer will be
created.

The example below will create an NGINX deployment and expose it via a load
balancer serving http on port 80, and https on 443. Note that the service
**type** is set to **LoadBalancer**.

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
    service.beta.kubernetes.io/oci-load-balancer-ssl-ports: "443"
    service.beta.kubernetes.io/oci-load-balancer-tls-secret: ssl-certificate-secret
spec:
  selector:
    app: nginx
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: 80
  - name: https
    port: 443
    targetPort: 80
```

First the required secret needs to be created in Kubernetes. For the purposes
of this example, we will create a self-signed certificate. However, in
production you would most likely use a public certificate signed by a
certificate authority:

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=nginxsvc/O=nginxsvc"
kubectl create secret tls ssl-certificate-secret --key tls.key --cert tls.crt
```

Create the service:

```
$ kubectl create -f examples/nginx-demo-svc-ssl.yaml
```

Watch the service and await a public IP address. This will be the load balancer
IP which you can use to connect to your service.

```
$ kubectl get svc --watch
NAME            CLUSTER-IP     EXTERNAL-IP      PORT(S)        AGE
nginx-service   10.96.97.137   129.213.12.174   80:30274/TCP   5m
```

You can now access your service via the provisioned load balancer using either
http or https:

```
curl http://129.213.12.174
curl --insecure https://129.213.12.174
```

Note: The `--insecure` flag above is only required due to our use of self-signed
certificates in this example.
