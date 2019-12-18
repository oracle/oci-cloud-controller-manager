module github.com/oracle/oci-cloud-controller-manager

go 1.12

replace (
	bitbucket.oci.oraclecorp.com/oke/oke-common => bitbucket.oci.oraclecorp.com/oke/oke-common v1.0.1-0.20190917222423-ba5e028f261d
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2
	github.com/oracle/oci-go-sdk => bitbucket.oci.oraclecorp.com/sdk/oci-go-sdk v0.0.0-00000000000000-e5ddf2b284c
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e
	oracle.com/oci/httpiam => bitbucket.oci.oraclecorp.com/goiam/httpiam.git v0.0.0-00000000000000-973dbb679788e9727a86d30ea0cccadcc0fe33d6 // 0.14.
	oracle.com/oci/httpsigner => bitbucket.oci.oraclecorp.com/goiam/httpsigner.git v0.0.0-00000000000000-e8cb27ebf4409946b295b9e22e511a52fc967e91 // 0.17.1
	oracle.com/oci/ociauthz => bitbucket.oci.oraclecorp.com/goiam/ociauthz.git v0.0.0-00000000000000-b00a4280e2092ac2c220111731965f49392734c1
	oracle.com/oci/ocihttpiam => bitbucket.oci.oraclecorp.com/goiam/ocihttpiam.git v0.0.0-00000000000000-996aa4a919d9e80238807c1c63c385980a0302a8
	oracle.com/oci/tagging => bitbucket.oci.oraclecorp.com/GOPLEX/tagging.git v0.0.0-00000000000000-20a2e48911da14e503935718f66588ab14aad8d4
	oracle.com/oke/oci-go-common => bitbucket.oci.oraclecorp.com/oke/oci-go-common.git v0.0.0-00000000000000-f93927b2b66cb1de2a10cf0f9f0d7e349bc0ae27

)

require (
	bitbucket.oci.oraclecorp.com/oke/bmc-go-sdk v0.0.0-20180119170638-a7c726955dd4 // indirect
	bitbucket.oci.oraclecorp.com/oke/oke-common v1.0.1-0.20190917222423-ba5e028f261d
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/Sirupsen/logrus v1.4.2 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/container-storage-interface/spec v1.1.0
	github.com/docker/distribution v0.0.0-20180720172123-0dae0957e5fe // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/mock v1.2.0 // indirect
	github.com/golang/protobuf v1.3.1
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.6.1
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible // indirect
	github.com/kubernetes-csi/external-attacher v2.0.0+incompatible
	github.com/kubernetes-csi/external-provisioner v1.4.0
	github.com/kubernetes-csi/external-snapshotter v1.0.1
	github.com/miekg/dns v1.1.17 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/oracle/oci-go-sdk v0.0.0-00010101000000-000000000000
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_golang v0.9.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0 // indirect
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20190904154756-749cb33beabd
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/genproto v0.0.0-20190418145605-e7d98fc518a7 // indirect
	google.golang.org/grpc v1.19.1
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/cloud-provider v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/csi-translation-lib v0.0.0
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.15.0
	k8s.io/utils v0.0.0-20190907131718-3d4f5b7dea0b
	oracle.com/oci/httpsigner v0.0.0-00010101000000-000000000000 // indirect
	oracle.com/oci/ociauthz v0.0.0-00010101000000-000000000000 // indirect
	oracle.com/oci/tagging v0.0.0-00010101000000-000000000000 // indirect
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.0.1+incompatible
)
