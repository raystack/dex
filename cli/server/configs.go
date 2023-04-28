package server

import (
	"fmt"

	"github.com/goto/salt/config"
	"github.com/spf13/cobra"

	"github.com/goto/dex/pkg/errors"
	"github.com/goto/dex/pkg/logger"
	"github.com/goto/dex/pkg/telemetry"
)

// serverConfig contains the application configuration.
type serverConfig struct {
	Log         logger.LogConfig `mapstructure:"log"`
	Service     serveConfig      `mapstructure:"service"`
	Shield      shieldConfig     `mapstructure:"shield"`
	Entropy     entropyConfig    `mapstructure:"entropy"`
	Siren       sirenConfig      `mapstructure:"siren"`
	Telemetry   telemetry.Config `mapstructure:"telemetry"`
	Odin        odinConfig       `mapstructure:"odin"`
	Compass     compassConfig    `mapstructure:"compass"`
	StencilAddr string           `mapstructure:"stencil_addr"`
}

type odinConfig struct {
	Addr string `mapstructure:"addr"`
}

type shieldConfig struct {
	Addr string `mapstructure:"addr"`
}

type entropyConfig struct {
	Addr string `mapstructure:"addr"`
}

type sirenConfig struct {
	Addr string `mapstructure:"addr"`
}

type compassConfig struct {
	Addr string `mapstructure:"addr"`
}

type serveConfig struct {
	Host string `mapstructure:"host" default:""`
	Port int    `mapstructure:"port" default:"8080"`
}

func (serveCfg serveConfig) Addr() string {
	return fmt.Sprintf("%s:%d", serveCfg.Host, serveCfg.Port)
}

func loadConfig(cmd *cobra.Command) (serverConfig, error) {
	configFile, _ := cmd.Flags().GetString("config")
	loader := config.NewLoader(config.WithFile(configFile))

	var cfg serverConfig
	if err := loader.Load(&cfg); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			fmt.Println(err)
			return cfg, nil
		}
		return serverConfig{}, err
	}

	return cfg, nil
}
