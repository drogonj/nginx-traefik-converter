package cmd

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/spf13/cobra"
	"log/slog"
)

// Config holds the information of the cli config.
type Config struct {
	NoColor     bool
	LogLevel    string
	IngressFile string
	ToFile      string
	Files       []string
}

var (
	cliCfg = new(Config)
	opts   = configs.NewOptions()
	logger *slog.Logger
)

// Registers all global flags to utility.
func registerCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cliCfg.LogLevel, "log-level", "", "INFO",
		"log level for the ingress-traefik-converter")
	cmd.PersistentFlags().StringVarP(&cliCfg.IngressFile, "ingress-file", "", "",
		"path to ingress file")
	cmd.PersistentFlags().StringArrayVarP(&cliCfg.Files, "file", "f", nil,
		"root yaml files to be used for importing")
	cmd.PersistentFlags().BoolVarP(&cliCfg.NoColor, "no-color", "", false,
		"when enabled the output would not be color encoded")
	cmd.PersistentFlags().BoolVarP(&opts.ProxyBufferHeuristic, "proxy-buffer-heuristic", "", false,
		"when enabled, the nginx ingress annotation 'proxy-buffer-size' gets heuristically mapped to Traefik buffering")
}

func registerImportFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cliCfg.ToFile, "to-file", "", "",
		"name of the file to which the final imported yaml should be written to")
}
