package cmd

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/log"
	"github.com/spf13/cobra"
)

func setCLIClient(_ *cobra.Command, _ []string) error {
	//writer = os.Stdout
	//
	//if len(cliCfg.ToFile) != 0 {
	//	filePTR, err := os.Create(cliCfg.ToFile)
	//	if err != nil {
	//		return err
	//	}
	//
	//	writer = filePTR
	//}

	logger = log.SetLogger(cliCfg.LogLevel)

	return nil
}
