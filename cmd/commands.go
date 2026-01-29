package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/convert"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/ingress"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/render"
	"log/slog"
	"os"
	"strings"

	"github.com/nikhilsbhat/ingress-traefik-converter/version"
	"github.com/spf13/cobra"
)

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "ingress-traefik-converter [command]",
		Short: "A utility to facilitate the conversion of nginx ingress to traefik.",
		Long:  `It identifies the nginx ingress present in the system and converts them to traefik equivalents.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Usage()
		},
	}
	rootCommand.SetUsageTemplate(getUsageTemplate())

	return rootCommand
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of ingress-traefik-converter installed",
		Long:  `This will help the user find what version of the ingress-traefik-converter he or she installed in her machine.`,
		RunE:  versionConfig,
	}
}

func getImportCommand() *cobra.Command {
	importCommand := &cobra.Command{
		Use:     "convert [flags]",
		Short:   "Converts the ingress nginx to equivalent trafik configs",
		Long:    "Command that reads the existing nginx ingress and creates an alternatives in traefik, it auto maps annotations",
		Example: ``,
		PreRunE: setCLIClient,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ing, err := ingress.Load(cliCfg.IngressFile)
			if err != nil {
				logger.Error("loading ingress errored",
					slog.Any("ingress", ing.Name),
					slog.Any("error:", err.Error()))

				return err
			}

			res := configs.NewResult()
			ctx := configs.New(ing, res)

			if err = convert.Run(*ctx, *opts); err != nil {
				logger.Error("converting ingress to traefik errored",
					slog.Any("ingress", ing.Name),
					slog.Any("error:", err.Error()))
				return err
			}

			if err = render.WriteYAML(*res, "./out"); err != nil {
				logger.Error("writing converted traefik ingress errored",
					slog.Any("ingress", ing.Name),
					slog.Any("error:", err.Error()))

				return err
			}

			return nil
		},
	}

	importCommand.SilenceErrors = true
	registerCommonFlags(importCommand)
	registerImportFlags(importCommand)

	return importCommand
}

func versionConfig(_ *cobra.Command, _ []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		logger.Error("version fetch of yaml failed", slog.Any("err", err))
		os.Exit(1)
	}

	versionWriter := bufio.NewWriter(os.Stdout)
	versionInfo := fmt.Sprintf("%s \n", strings.Join([]string{"yamll version", string(buildInfo)}, ": "))

	if _, err = versionWriter.WriteString(versionInfo); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}(versionWriter)

	return nil
}
