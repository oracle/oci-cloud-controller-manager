module github.com/oracle/oci-cloud-controller-manager

go 1.12

replace (
	github.com/docker/docker => github.com/docker/engine v0.0.0-20181106193140-f5749085e9cb
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
	k8s.io/component-base => k8s.io/component-base v0.17.0
	k8s.io/cri-api => k8s.io/cri-api v0.16.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.0
	k8s.io/klog => github.com/mrunalpagnis/klog v0.0.0-00000000000000-ec66c0a95a3fe542357d0366ad25f152cce66b8b
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.4
	k8s.io/kubectl => k8s.io/kubectl v0.16.4
	k8s.io/kubelet => k8s.io/kubelet v0.16.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.4
	k8s.io/metrics => k8s.io/metrics v0.16.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.4
)

require (
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/container-storage-interface/spec v1.2.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.7.1
	github.com/kubernetes-csi/external-attacher v2.1.0+incompatible // indirect
	github.com/kubernetes-csi/external-provisioner v1.6.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/oracle/oci-go-sdk v0.0.0-00000000000000-857a54193bc062afe3933d02b38e3fe9d158f744
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.4.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/zap v1.9.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/grpc v1.29.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/cloud-provider v0.17.0
	k8s.io/component-base v0.17.0
	k8s.io/klog v1.0.0
	k8s.io/kubectl v1.16.4 // indirect
	k8s.io/kubernetes v1.16.0
	k8s.io/utils v0.0.0-20200124190032-861946025e34
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.1.0+incompatible
)
