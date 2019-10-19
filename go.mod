module istio.io/operator

go 1.13

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

require (
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.14.1+incompatible // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/frankban/quicktest v1.4.1 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/gorilla/mux v1.7.2 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/kr/pretty v0.1.0
	github.com/kylelemons/godebug v1.1.0
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/openshift/api v3.9.1-0.20191008181517-e4fd21196097+incompatible // indirect
	github.com/openshift/cluster-network-operator v0.0.0-20191009144453-fdceef8e1a7b
	github.com/pierrec/lz4 v2.2.5+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/prom2json v1.2.1 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.3.0
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	istio.io/pkg v0.0.0-20190515193414-9332430ad747
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/component-base v0.0.0-20191016230640-d338b9159fb6 // indirect
	k8s.io/helm v2.14.3+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-proxy v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/kubectl v0.0.0-20191003004222-1f3c0cd90ca9
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/yaml v1.1.0
)

replace k8s.io/api => k8s.io/api v0.0.0-20191003000013-35e20aa79eb8

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191003003426-b4b1f434fead

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191003000419-f68efa97b39e

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191003001037-3c8b233e046c

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191003002041-49e3d608220c

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191003002833-e367e4712542

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191003002408-6e42c232ac7d

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191003003129-09316795c0dd

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190927045949-f81bca4f5e85

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191003003001-314f0beee0a9

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191003002707-f6b7b0f55cc0

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191003003551-0eecdcdcc049

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191003003732-7d49cdad1c12

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20191003000551-f573d376509c

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191003003255-c493acd9e2ff

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20191003002233-837aead57baf

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191003001538-80f33ca02582

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191003001317-a019a9d85a86
