package webserver

import (
	"net"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	cch "github.com/patrickmn/go-cache"
)

// Config is parameter that used for starting web server
type Config struct {
	DB            model.DB
	DataDir       string
	ServerAddress string
	ServerPort    int
	RootPath      string
	Log           bool
}

// GetLegacyHandler returns a legacy handler to use with the new webserver
func GetLegacyHandler(cfg Config, dependencies model.Dependencies) *Handler {
	plainIPs := dependencies.Config().Http.SSOProxyAuthTrusted
	trustedIPs := make([]*net.IPNet, len(plainIPs))
	for i, ip := range plainIPs {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			dependencies.Logger().WithError(err).WithField("ip", ip).Error("Failed to parse trusted ip cidr")
			continue
		}

		trustedIPs[i] = ipNet
	}

	return &Handler{
		DB:        cfg.DB,
		DataDir:   cfg.DataDir,
		UserCache: cch.New(time.Hour, 10*time.Minute),
		// SessionCache: cch.New(time.Hour, 10*time.Minute),
		ArchiveCache: cch.New(time.Minute, 5*time.Minute),
		RootPath:     cfg.RootPath,
		Log:          cfg.Log,
		dependencies: dependencies,
		trustedIPs:   trustedIPs,
	}
}
