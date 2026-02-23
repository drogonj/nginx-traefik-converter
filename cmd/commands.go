package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/convert"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/render"
	"github.com/nikhilsbhat/nginx-traefik-converter/version"
	"github.com/spf13/cobra"
)

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "nginx-traefik-converter [command]",
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
		Short: "Command to fetch the version of nginx-traefik-converter installed",
		Long:  `This will help the user find what version of the nginx-traefik-converter he or she installed in her machine.`,
		RunE:  versionConfig,
	}
}

func getConvertCommand() *cobra.Command {
	convertCommand := &cobra.Command{
		Use:     "convert [flags]",
		Short:   "Converts the ingress nginx to equivalent trafik configs",
		Long:    "Command that reads the existing nginx ingress and creates an alternatives in traefik, it auto maps annotations",
		Example: ``,
		PreRunE: setCLIClient,
		RunE: func(_ *cobra.Command, _ []string) error {
			ingresses, err := kubeConfig.ListAllIngresses()
			if err != nil {
				return err
			}

			var globalReport configs.GlobalReport

			for _, ingress := range ingresses {
				res := configs.NewResult()
				ctx := configs.New(&ingress, res, opts, logger)
				ctx.StartIngressReport(ingress.Namespace, ingress.Name)

				if err = convert.Run(*ctx); err != nil {
					logger.Error("converting ingress to traefik errored",
						slog.String("ingress", ingress.Name),
						slog.String("error", err.Error()))

					continue
				}

				if err = render.WriteYAML(*res, filepath.Join("./out", ingress.Name)); err != nil {
					logger.Error("writing converted traefik ingress errored",
						slog.String("ingress", ingress.Name),
						slog.String("error", err.Error()))

					return err
				}

				if err = printerConfig.PrintIngressSummary(ctx.Result.IngressReport); err != nil {
					return err
				}

				globalReport.Ingresses = append(
					globalReport.Ingresses,
					ctx.Result.IngressReport,
				)
			}

			if err = printerConfig.PrintGlobalSummary(globalReport); err != nil {
				return err
			}

			logger.Info("nginx ingress to traefik conversion completed")

			return nil
		},
	}

	convertCommand.SilenceErrors = true
	registerCommonFlags(convertCommand)
	registerImportFlags(convertCommand)

	return convertCommand
}

func getSupportedAnnotationCommand() *cobra.Command {
	supportedAnnotationsCommand := &cobra.Command{
		Use:     "supported-annotations [flags]",
		Short:   "list supported annotaions",
		Long:    "Command list all the annotations that converter supports",
		Example: ``,
		PreRunE: setCLIClient,
		RunE: func(_ *cobra.Command, _ []string) error {
			annotations := models.GetAnnotations()

			for _, annotation := range annotations {
				fmt.Printf("%s\n", annotation)
			}

			return nil
		},
	}

	supportedAnnotationsCommand.SilenceErrors = true
	// registerCommonFlags(supportedAnnotationsCommand)
	// registerImportFlags(supportedAnnotationsCommand)

	return supportedAnnotationsCommand
}

func versionConfig(_ *cobra.Command, _ []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		logger.Error("version fetch of yaml failed", slog.Any("err", err))
		os.Exit(1)
	}

	versionWriter := bufio.NewWriter(os.Stdout)
	versionInfo := fmt.Sprintf("%s \n", strings.Join([]string{"nginx-traefik-converter version", string(buildInfo)}, ": "))

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
