package config

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

// readDotEnv reads the configuration from variables in a .env file (only for contributing)
func readDotEnv(logger *zap.Logger) map[string]string {
	file, err := os.Open(".env")
	if err != nil {
		return nil
	}
	defer file.Close()

	result := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		keyval := strings.SplitN(line, "=", 2)
		result[keyval[0]] = keyval[1]
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal("error reading dotenv", zap.Error(err))
	}

	return result
}

type HttpConfig struct {
	Enabled   bool   `env:"HTTP_ENABLED,default=True"`
	Port      int    `env:"HTTP_PORT,default=8080"`
	Address   string `env:"HTTP_ADDRESS,default=:"`
	RootPath  string `env:"HTTP_ROOT_PATH,default=/"`
	AccessLog bool   `env:"HTTP_ACCESS_LOG,default=True"`
	SecretKey string `env:"HTTP_SECRET_KEY"`
	// Fiber Specific
	BodyLimit                    int           `env:"HTTP_BODY_LIMIT,default=1024"`
	ReadTimeout                  time.Duration `env:"HTTP_READ_TIMEOUT,default=10s"`
	WriteTimeout                 time.Duration `env:"HTTP_WRITE_TIMEOUT,default=10s"`
	IDLETimeout                  time.Duration `env:"HTTP_IDLE_TIMEOUT,default=10s"`
	DisableKeepAlive             bool          `env:"HTTP_DISABLE_KEEP_ALIVE,default=true"`
	DisablePreParseMultipartForm bool          `env:"HTTP_DISABLE_PARSE_MULTIPART_FORM,default=true"`
	Routes                       struct {
		Bookmark struct {
			Path string `env:"ROUTES_BOOKMARK_PATH,default=/bookmark"`
		}
		Frontend struct {
			Path   string        `env:"ROUTES_STATIC_PATH,default=/"`
			MaxAge time.Duration `env:"ROUTES_STATIC_MAX_AGE,default=720h"`
		}
		System struct {
			Path string `env:"ROUTES_SYSTEM_PATH,default=/system"`
		}
		API struct {
			Path string `env:"ROUTE_API_PATH,default=/api/v1"`
		}
	}
	Storage struct {
		DataDir string `env:"DATA_DIR"`
	}
}

type Config struct {
	Hostname string `env:"HOSTNAME,required"`
	// LogLevel string `env:"LOG_LEVEL,default=info"`
	Http HttpConfig
}

func ParseServerConfiguration(ctx context.Context, logger *zap.Logger) *Config {
	var cfg Config

	lookuper := envconfig.MultiLookuper(
		envconfig.MapLookuper(map[string]string{"HOSTNAME": os.Getenv("HOSTNAME")}),
		envconfig.MapLookuper(readDotEnv(logger)),
		envconfig.PrefixLookuper("SHIORI_", envconfig.OsLookuper()),
		envconfig.OsLookuper(),
	)
	if err := envconfig.ProcessWith(ctx, &cfg, lookuper); err != nil {
		logger.Fatal("Error parsing configuration: %s", zap.Error(err))
	}

	return &cfg
}
