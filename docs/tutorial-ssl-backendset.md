# Tutorial

This example will show you how to use the CCM to create a load balancer with SSL
termination for BackendSets.

### Load balancer with SSL termination for BackendSets example

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
    service.beta.kubernetes.io/oci-load-balancer-tls-backendset-secret: ssl-certificate-secret
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

First, the required Secret needs to be created in Kubernetes. For the purposes
of this example, we will create a self-signed certificate. However, in
production you would most likely use a public certificate signed by a
certificate authority.

Below is an example of a secret configuration file required to be uploaded as a Kubernetes
generic Secret. The CA certificate, the public certificate and the private key need to be base64 encoded:

***Note: Certificates for BackendSets require a CA certificate to be provided.***

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ssl-certificate-secret
type: Opaque
data:
  ca.crt: LS0tLS1CRUdJTiBDRV(...)
  tls.crt: LS0tLS1CRUdJTi(...)
  tls.key: LS0tLS1CRUdJTi(...)
```

```
kubectl create -f ssl-certificate-secret.yaml
```

Create the service:

```
$ kubectl create -f manifests/demo/nginx-demo-svc-ssl.yaml
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
