# Install the CCM on Ubuntu

This guide will explain how to manually install and setup the CCM on a vanilla
Ubuntu 16.04 cluster created with kubeadm.

**Note the you will need a VCN with security rules that allow UDP and TCP
traffic. Flannel requires UDP access.**

### Prerequisites

 - Create two (or more) OCI Ubuntu 16.04 instances. One master, N nodes.

## Install Kubernetes with kubeadm

Run the following script on both the master and worker nodes.

**Run the following as root (sudo su - or similar)**

```bash
apt-get install apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu xenial stable"
apt-get update

apt-get install -y docker-ce
usermod -aG docker $USER

apt-get update && apt-get install -y apt-transport-https
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF

apt-get update
apt-get install -y kubelet kubeadm
```

Run the following on the master only as a normal user

```bash
sudo iptables -A INPUT -p tcp --dport 6443 -j ACCEPT && sudo iptables -F

EXTERNAL_IP=$(curl ipinfo.io/ip)

sudo kubeadm init \
--apiserver-cert-extra-sans=${EXTERNAL_IP} \
--pod-network-cidr=10.244.0.0/16 \
--apiserver-advertise-address=${EXTERNAL_IP}

kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel-rbac.yml
```

## Join the nodes to the master

Run on all nodes

```
sudo kubeadm join --token XXX MASTER_IP:6443 --skip-preflight-checks
```

## Update kubelet on all nodes (including master)

Add the following to `/etc/systemd/system/kubelet.service.d/10-kubeadm.conf`

```
Environment=KUBELET_EXTRA_ARGS="--cloud-provider=external"
```

Restart kubelet

```
sudo systemctl daemon-reload
```

## Update kube-controller-manager on master

```
sudo sed -i '/    - kube-controller-manager/a\\    - --cloud-provider=external' /etc/kubernetes/manifests/kube-controller-manager.yaml
```

Your cluster is now configured to use an external cloud provider.

## Finish up

It's important to note that you may have issues with the Ubuntu firewall. You'll
need to disable it or update IP table rules.

```
sudo ufw disable
```
