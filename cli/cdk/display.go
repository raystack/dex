package cdk

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yudai/pp"
	"gopkg.in/yaml.v2"

	"github.com/odpf/dex/pkg/errors"
)

type FormatFn func(w io.Writer, v interface{}) error

func Display(cmd *cobra.Command, v interface{}, prettyFormatter FormatFn) error {
	format, _ := cmd.Flags().GetString("format")
	format = strings.ToLower(strings.TrimSpace(format))

	var formatter FormatFn
	switch format {
	case "json":
		formatter = JSONFormat

	case "yaml", "yml":
		formatter = YAMLFormat

	case "pretty", "human":
		if prettyFormatter != nil {
			formatter = prettyFormatter
		} else {
			formatter = PPFormatter
		}

	case "pp":
		formatter = PPFormatter
	}

	if formatter == nil {
		return errors.Errorf("--format value '%s' is not valid", format)
	}

	return formatter(os.Stdout, v)
}

func JSONFormat(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func YAMLFormat(w io.Writer, v interface{}) error {
	enc := yaml.NewEncoder(w)
	return enc.Encode(v)
}

func PPFormatter(w io.Writer, v interface{}) error {
	_, err := pp.Fprintln(w, v)
	return err
}
