package config

import (
	"fmt"

	"github.com/odpf/dex/pkg/logger"
	"github.com/odpf/dex/pkg/telemetry"
)

// Config contains the application configuration.
type Config struct {
	Log       logger.LogConfig `mapstructure:"log"`
	Service   serveConfig      `mapstructure:"service"`
	Telemetry telemetry.Config `mapstructure:"telemetry"`
}

type serveConfig struct {
	Host string `mapstructure:"host" default:""`
	Port int    `mapstructure:"port" default:"8080"`
}

func (serveCfg serveConfig) Addr() string {
	return fmt.Sprintf("%s:%d", serveCfg.Host, serveCfg.Port)
}
