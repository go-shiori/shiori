package cmd

import (
	"testing"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func Test_setIfFlagChanged(t *testing.T) {
	type args struct {
		flagName string
		flags    func() *pflag.FlagSet
		cfg      *config.Config
		fn       func(cfg *config.Config)
	}
	tests := []struct {
		name     string
		args     args
		assertFn func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "Flag didn't change",
			args: args{
				flagName: "port",
				flags: func() *pflag.FlagSet {
					return &pflag.FlagSet{}
				},
				cfg: &config.Config{
					Http: &config.HttpConfig{
						Port: 8080,
					},
				},
				fn: func(cfg *config.Config) {
					cfg.Http.Port = 9999
				},
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				require.Equal(t, cfg.Http.Port, 8080)
			},
		},
		{
			name: "Flag changed",
			args: args{
				flagName: "port",
				flags: func() *pflag.FlagSet {
					pf := &pflag.FlagSet{}
					pf.IntP("port", "p", 8080, "Port used by the server")
					pf.Set("port", "9999")
					return pf
				},
				cfg: &config.Config{
					Http: &config.HttpConfig{
						Port: 8080,
					},
				},
				fn: func(cfg *config.Config) {
					cfg.Http.Port = 9999
				},
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				require.Equal(t, cfg.Http.Port, 9999)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setIfFlagChanged(tt.args.flagName, tt.args.flags(), tt.args.cfg, tt.args.fn)
		})
	}
}
