package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/odpf/dex/pkg/errors"
	"github.com/odpf/dex/pkg/logger"
	"github.com/odpf/dex/pkg/telemetry"
)

const configFlag = "config"

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

func (serveCfg serveConfig) addr() string {
	return fmt.Sprintf("%s:%d", serveCfg.Host, serveCfg.Port)
}

func cmdShowConfigs() *cobra.Command {
	return &cobra.Command{
		Use:   "configs",
		Short: "Display configurations currently loaded",
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				fatalExitf("failed to read configs: %v", err)
			}
			return yaml.NewEncoder(os.Stdout).Encode(cfg)
		}),
	}
}

func loadConfig(cmd *cobra.Command) (Config, error) {
	var opts []config.LoaderOption

	cfgFile, _ := cmd.Flags().GetString(configFlag)
	if cfgFile != "" {
		opts = append(opts, config.WithFile(cfgFile))
	} else {
		opts = append(opts,
			config.WithPath("./"),
			config.WithName("dex"),
		)
	}

	var cfg Config
	err := config.NewLoader(opts...).Load(&cfg)
	if errors.As(err, &config.ConfigFileNotFoundError{}) {
		log.Println(err)
	} else {
		return cfg, err
	}

	return cfg, nil
}
