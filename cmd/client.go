package cmd

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/log"
	"github.com/spf13/cobra"
)

func setCLIClient(_ *cobra.Command, _ []string) error {
	logger = log.SetLogger(cliCfg.LogLevel)

	kubeConfig.SetLogger(logger)

	if err := kubeConfig.SetKubeClient(); err != nil {
		return err
	}

	kubeConfig.SetKubeNameSpace()

	return nil
}
