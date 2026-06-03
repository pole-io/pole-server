module github.com/pole-io/pole-server

go 1.25

require (
	github.com/BurntSushi/toml v1.2.0
	github.com/emicklei/go-restful/v3 v3.9.0
	github.com/envoyproxy/go-control-plane v0.13.4
	github.com/go-openapi/spec v0.20.7 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mitchellh/mapstructure v1.4.3
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/nicksnyder/go-i18n/v2 v2.2.0
	github.com/pkg/errors v0.9.1
	github.com/polarismesh/go-restful-openapi/v2 v2.0.0-20220928152401-083908d10219
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.10.0
	go.uber.org/atomic v1.10.0
	go.uber.org/automaxprocs v1.4.0
	go.uber.org/zap v1.23.0
	golang.org/x/crypto v0.38.0
	golang.org/x/net v0.40.0
	golang.org/x/sync v0.14.0
	golang.org/x/text v0.25.0
	golang.org/x/time v0.1.1-0.20221020023724-80b9fac54d29
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.6
)

// Indirect dependencies group
require (
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/envoyproxy/go-control-plane/envoy v1.32.4
	github.com/gin-gonic/gin v1.4.0
	github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/mark3labs/mcp-go v0.18.0
	github.com/pole-io/specification v0.1.0-ALPHA.23
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.36.0
	go.opentelemetry.io/otel/metric v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cel.dev/expr v0.20.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/envoyproxy/go-control-plane/ratelimit v0.1.0 // indirect
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/ugorji/go v1.1.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
)

require (
	github.com/dlclark/regexp2 v1.10.0
	go.etcd.io/bbolt v1.3.7
	go.opentelemetry.io/contrib/instrumentation/runtime v0.61.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237 // indirect
)
