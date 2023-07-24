package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
)

// readDotEnv reads the configuration from variables in a .env file (only for contributing)
func readDotEnv(logger *logrus.Logger) map[string]string {
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
		logger.WithError(err).Fatal("error reading dotenv")
	}

	return result
}

type HttpConfig struct {
	Enabled    bool   `env:"HTTP_ENABLED,default=True"`
	Port       int    `env:"HTTP_PORT,default=8080"`
	Address    string `env:"HTTP_ADDRESS,default=:"`
	RootPath   string `env:"HTTP_ROOT_PATH,default=/"`
	AccessLog  bool   `env:"HTTP_ACCESS_LOG,default=True"`
	ServeWebUI bool   `env:"HTTP_SERVE_WEB_UI,default=True"`
	SecretKey  string `env:"HTTP_SECRET_KEY"`
	// Fiber Specific
	BodyLimit                    int           `env:"HTTP_BODY_LIMIT,default=1024"`
	ReadTimeout                  time.Duration `env:"HTTP_READ_TIMEOUT,default=10s"`
	WriteTimeout                 time.Duration `env:"HTTP_WRITE_TIMEOUT,default=10s"`
	IDLETimeout                  time.Duration `env:"HTTP_IDLE_TIMEOUT,default=10s"`
	DisableKeepAlive             bool          `env:"HTTP_DISABLE_KEEP_ALIVE,default=true"`
	DisablePreParseMultipartForm bool          `env:"HTTP_DISABLE_PARSE_MULTIPART_FORM,default=true"`
}

type DatabaseConfig struct {
	DBMS string `env:"DBMS"` // Deprecated
	// DBMS requires more environment variables. Check the database package for more information.
	URL string `env:"DATABASE_URL"`
}

type StorageConfig struct {
	DataDir string `env:"DIR"` // Using DIR to be backwards compatible with the old config
}

type Config struct {
	Hostname    string `env:"HOSTNAME,required"`
	Development bool   `env:"DEVELOPMENT,default=False"`
	Database    *DatabaseConfig
	Storage     *StorageConfig
	// LogLevel string `env:"LOG_LEVEL,default=info"`
	Http *HttpConfig
}

// SetDefaults sets the default values for the configuration
func (c *HttpConfig) SetDefaults(logger *logrus.Logger) {
	// Set a random secret key if not set
	if c.SecretKey == "" {
		logger.Warn("SHIORI_HTTP_SECRET_KEY is not set, using random value. This means that all sessions will be invalidated on server restart.")
		randomUUID, err := uuid.NewV4()
		if err != nil {
			logger.WithError(err).Fatal("couldn't generate a random UUID")
		}
		c.SecretKey = randomUUID.String()
	}
}

// SetDefaults sets the default values for the configuration
func (c Config) SetDefaults(logger *logrus.Logger, portableMode bool) {
	// Set the default storage directory if not set, setting also the database url for
	// sqlite3 if that engine is used
	if c.Storage.DataDir == "" {
		var err error
		c.Storage.DataDir, err = getStorageDirectory(portableMode)
		if err != nil {
			logger.WithError(err).Fatal("couldn't determine the data directory")
		}
	}

	// Set default database url if not set
	if c.Database.DBMS == "" && c.Database.URL == "" {
		c.Database.URL = fmt.Sprintf("sqlite:///%s", filepath.Join(c.Storage.DataDir, "shiori.db"))
	}
}

func ParseServerConfiguration(ctx context.Context, logger *logrus.Logger) *Config {
	var cfg Config

	lookuper := envconfig.MultiLookuper(
		envconfig.MapLookuper(map[string]string{"HOSTNAME": os.Getenv("HOSTNAME")}),
		envconfig.MapLookuper(readDotEnv(logger)),
		envconfig.PrefixLookuper("SHIORI_", envconfig.OsLookuper()),
		envconfig.OsLookuper(),
	)
	if err := envconfig.ProcessWith(ctx, &cfg, lookuper); err != nil {
		logger.WithError(err).Fatal("Error parsing configuration")
	}

	return &cfg
}
