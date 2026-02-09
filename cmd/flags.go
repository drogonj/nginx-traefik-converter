package cmd

import (
	"log/slog"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/kubernetes"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/render"
	"github.com/spf13/cobra"
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
	cliCfg        = new(Config)
	opts          = configs.NewOptions()
	logger        *slog.Logger
	kubeConfig    = kubernetes.New()
	printerConfig = render.New()
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
	cmd.PersistentFlags().StringVarP(&kubeConfig.Context, "context", "c", "",
		"kubernetes context to use")
	cmd.PersistentFlags().StringVarP(&kubeConfig.NameSpace, "namespace", "n", "default",
		"kubernetes namespace to set")
	cmd.PersistentFlags().BoolVarP(&kubeConfig.All, "all", "a", false,
		"when set, all namespaces would be considered")
}

func registerImportFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cliCfg.ToFile, "to-file", "", "",
		"name of the file to which the final imported yaml should be written to")
	cmd.PersistentFlags().BoolVarP(&printerConfig.Table, "table", "", false,
		"when enabled prints output in table format")
}
