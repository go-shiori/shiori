module github.com/go-shiori/shiori

// +heroku goVersion go1.22
go 1.23

toolchain go1.23.4

require (
	git.sr.ht/~emersion/go-sqlite3-fts5 v0.0.0-20240124102820-f3a72e8b79b1
	github.com/PuerkitoBio/goquery v1.10.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/fatih/color v1.18.0
	github.com/gin-contrib/requestid v1.0.4
	github.com/gin-contrib/static v1.1.3
	github.com/gin-gonic/gin v1.10.0
	github.com/go-shiori/go-epub v1.2.2-0.20240211121944-dc6435eac436
	github.com/go-shiori/go-readability v0.0.0-20241012063810-92284fa8a71f
	github.com/go-shiori/warc v0.0.0-20200621032813-359908319d1d
	github.com/go-sql-driver/mysql v1.8.1
	github.com/gofrs/uuid/v5 v5.3.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/muesli/go-app-paths v0.2.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-envconfig v1.1.0
	github.com/shurcooL/vfsgen v0.0.0-20230704071429-0000e147ea92
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/afero v1.11.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.4
	github.com/testcontainers/testcontainers-go v0.34.0
	github.com/toorop/gin-logrus v0.0.0-20210225092905-2c785434f26f
	golang.org/x/crypto v0.31.0
	golang.org/x/image v0.23.0
	golang.org/x/net v0.33.0
	golang.org/x/term v0.27.0
	modernc.org/sqlite v1.34.4
)

require (
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.12.9 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/bytedance/sonic v1.12.6 // indirect
	github.com/bytedance/sonic/loader v0.2.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/containerd/containerd v1.7.24 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.4.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.7 // indirect
	github.com/gin-contrib/sse v1.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/go-shiori/dom v0.0.0-20230515143342-73569d674e1c // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20240909124753-873cd0166683 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/shurcooL/httpfs v0.0.0-20230704072500-f1e31cf0ba5c // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.9.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/bbolt v1.3.11 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	golang.org/x/arch v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20241217172543-b2144cdd0a67 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241230172942-26aa7a208def // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/gc/v3 v3.0.0-20241223112719-96e2e1e4408d // indirect
	modernc.org/libc v1.61.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/strutil v1.2.1 // indirect
	modernc.org/token v1.1.0 // indirect
)
