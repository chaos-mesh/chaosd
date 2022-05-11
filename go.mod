module github.com/chaos-mesh/chaosd

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/alecthomas/units v0.0.0-20210927113745-59d0afb8317a
	github.com/chaos-mesh/chaos-mesh v0.9.1-0.20220511035234-10df92fcde77
	github.com/chaos-mesh/chaos-mesh/api v0.0.0
	github.com/containerd/containerd v1.5.8
	github.com/docker/docker v20.10.7+incompatible
	github.com/gin-gonic/gin v1.7.2
	github.com/go-logr/zapr v1.2.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.2.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/joomcode/errorx v1.0.1
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/gomega v1.18.1
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/failpoint v0.0.0-20200210140405-f8f9fb234798
	github.com/pingcap/log v0.0.0-20200117041106-d28c14d3b1cd
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	github.com/swaggo/swag v1.6.7
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.40.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.7
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	sigs.k8s.io/controller-runtime v0.11.0
)

replace (
	// github.com/chaos-mesh/chaos-mesh require /api/v1alpha1 v0.0.0, but v0.0.0 can not be found, so use replace here
	github.com/chaos-mesh/chaos-mesh/api => github.com/chaos-mesh/chaos-mesh/api v0.0.0-20220511035234-10df92fcde77
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
	k8s.io/api => k8s.io/api v0.23.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.1
	k8s.io/apiserver => k8s.io/apiserver v0.23.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.1
	k8s.io/client-go => k8s.io/client-go v0.23.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.1
	k8s.io/code-generator => k8s.io/code-generator v0.23.1
	k8s.io/component-base => k8s.io/component-base v0.23.1
	k8s.io/cri-api => k8s.io/cri-api v0.23.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.1
	k8s.io/kubectl => k8s.io/kubectl v0.23.1
	k8s.io/kubelet => k8s.io/kubelet v0.23.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.1
	k8s.io/metrics => k8s.io/metrics v0.23.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.1
	vbom.ml/util => github.com/fvbommel/util v0.0.2
)

go 1.16
