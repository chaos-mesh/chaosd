module github.com/chaos-mesh/chaosd

require (
	github.com/chaos-mesh/chaos-mesh v0.9.1-0.20201223081802-ee3d6d987c92
	github.com/containerd/containerd v1.2.3
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/gin-gonic/gin v1.6.3
	github.com/google/uuid v1.1.1
	github.com/joomcode/errorx v1.0.1
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/gomega v1.9.0
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/failpoint v0.0.0-20200210140405-f8f9fb234798
	github.com/pingcap/log v0.0.0-20200117041106-d28c14d3b1cd
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v0.0.0-20180427012116-c95755e4bcd7
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/grpc v1.27.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.7
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.1-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.17.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.0
	k8s.io/code-generator => k8s.io/code-generator v0.17.1-beta.0
	k8s.io/component-base => k8s.io/component-base v0.17.0
	k8s.io/cri-api => k8s.io/cri-api v0.17.1-beta.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.0
	k8s.io/kubectl => k8s.io/kubectl v0.17.0
	k8s.io/kubelet => k8s.io/kubelet v0.17.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.0
	k8s.io/metrics => k8s.io/metrics v0.17.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.0
	vbom.ml/util => github.com/fvbommel/util v0.0.2
)

go 1.14
