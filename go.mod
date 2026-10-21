module github.com/chaos-mesh/chaosd

require (
	github.com/Jeffail/tunny v0.1.4
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b
	github.com/chaos-mesh/chaos-mesh v0.9.1-0.20261016030316-e020a2df5a69
	github.com/chaos-mesh/chaos-mesh/api v0.0.0
	github.com/containerd/containerd v1.7.36
	github.com/docker/docker v28.5.2+incompatible
	github.com/dustin/go-humanize v1.1.0
	github.com/gin-gonic/gin v1.12.0
	github.com/go-logr/logr v1.4.4
	github.com/go-logr/zapr v1.3.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/joomcode/errorx v1.2.0
	github.com/magiconair/properties v1.18.12
	github.com/moby/moby/client v0.6.2
	github.com/olekukonko/tablewriter v1.1.5
	github.com/onsi/gomega v1.44.0
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/log v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.25.0
	github.com/redis/go-redis/v9 v9.23.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/samber/lo v1.53.0
	github.com/samber/mo v1.17.0
	github.com/segmentio/kafka-go v0.4.51
	github.com/shirou/gopsutil/v4 v4.26.9
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.12.1
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	github.com/swaggo/swag v1.16.6
	go.uber.org/fx v1.24.0
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.84.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.2
	k8s.io/api v0.37.1
	k8s.io/apimachinery v0.37.1
	sigs.k8s.io/controller-runtime v0.25.2
)

require (
	cyphar.com/go-pathrs v0.2.6 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/AdamKorcz/go-118-fuzz-build v0.0.0-20250520111509-a70c2aa677fa // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.4 // indirect
	github.com/Microsoft/hcsshim v0.15.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.4 // indirect
	github.com/bytedance/sonic/loader v0.5.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chaos-mesh/chaos-driver v0.2.1 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/cgroups/v3 v3.1.3 // indirect
	github.com/containerd/containerd/api v1.12.0 // indirect
	github.com/containerd/continuity v0.5.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/log v0.2.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.5 // indirect
	github.com/containerd/ttrpc v1.2.10 // indirect
	github.com/containerd/typeurl/v2 v2.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/creachadair/jrpc2 v1.3.5 // indirect
	github.com/creachadair/mds v0.31.0 // indirect
	github.com/cyphar/filepath-securejoin v0.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.8.2 // indirect
	github.com/docker/go-events v0.1.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.11.2 // indirect
	github.com/emicklei/go-restful/v3 v3.14.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.6 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/gin-contrib/sse v1.1.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/go-openapi/jsonreference v1.0.3 // indirect
	github.com/go-openapi/spec v1.0.1 // indirect
	github.com/go-openapi/swag v0.29.2 // indirect
	github.com/go-openapi/swag/cmdutils v0.29.2 // indirect
	github.com/go-openapi/swag/conv v0.29.2 // indirect
	github.com/go-openapi/swag/fileutils v0.29.2 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/go-openapi/swag/loading v0.29.2 // indirect
	github.com/go-openapi/swag/mangling v0.29.2 // indirect
	github.com/go-openapi/swag/netutils v0.29.2 // indirect
	github.com/go-openapi/swag/pools v0.29.2 // indirect
	github.com/go-openapi/swag/stringutils v0.29.2 // indirect
	github.com/go-openapi/swag/typeutils v0.29.2 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.2 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/goccy/go-json v0.11.2 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/godbus/dbus/v5 v5.2.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/klauspost/cpuid/v2 v2.4.1 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260802145828-341c2f0c90b5 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-runewidth v0.0.31 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/moby/api v1.56.1 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.7.0 // indirect
	github.com/moby/sys/signal v0.7.1 // indirect
	github.com/moby/sys/user v0.4.1 // indirect
	github.com/moby/sys/userns v0.2.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
	github.com/olekukonko/errors v1.3.0 // indirect
	github.com/olekukonko/ll v0.1.9 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.3.0 // indirect
	github.com/opencontainers/selinux v1.15.1 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/pingcap/failpoint v0.0.0-20260811232634-55ac33a48e3b // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20260916203055-22a1a467d9f0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.64.0 // indirect
	github.com/retailnext/iptables_exporter v0.1.0 // indirect
	github.com/romana/ipset v1.0.0 // indirect
	github.com/romana/rlog v0.0.0-20220412051723-c08f605858a9 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.2 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.mongodb.org/mongo-driver/v2 v2.9.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/arch v0.32.0 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	k8s.io/apiextensions-apiserver v0.37.1 // indirect
	k8s.io/client-go v0.37.1 // indirect
	k8s.io/cri-api v0.37.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20261014044854-ca3452a79603 // indirect
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3 // indirect
	sigs.k8s.io/json v0.0.0-20260909141634-11ed52e25bc5 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

go 1.27.2

tool github.com/swaggo/swag/cmd/swag

replace (
	// Chaos Mesh requires an API version that is not published as a module tag.
	github.com/chaos-mesh/chaos-mesh/api => github.com/chaos-mesh/chaos-mesh/api v0.0.0-20261016030316-e020a2df5a69
	// The Chaos Mesh daemon still uses containerd 1.7 and its sandbox protocol.
	github.com/containerd/containerd/api => github.com/containerd/containerd/api v1.8.0
	// Upstream selects the obsolete v2 tag instead of the maintained v1 driver.
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.52
	github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v1.2.0
	// Kubernetes 0.37 uses structured-merge-diff v6 rather than v7.
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20260721132016-d427ff9ee9ad
)
