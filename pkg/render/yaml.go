package render

import (
	"fmt"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// WriteYAML writes the translated inputs to respective files.
func WriteYAML(res configs.Result, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	if err := writeObjects(
		filepath.Join(outDir, "middlewares.yaml"),
		res.Middlewares,
	); err != nil {
		return err
	}

	if err := writeObjects(
		filepath.Join(outDir, "ingressroutes.yaml"),
		res.IngressRoutes,
	); err != nil {
		return err
	}

	if err := writeObjects(
		filepath.Join(outDir, "tlsoptions.yaml"),
		res.TLSOptions); err != nil {
		return err
	}

	if len(res.Warnings) > 0 {
		if err := writeWarnings(
			filepath.Join(outDir, "warnings.txt"),
			res.Warnings,
		); err != nil {
			return err
		}
	}

	return nil
}

func writeObjects(path string, objs []client.Object) error {
	if len(objs) == 0 {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	for i, obj := range objs {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}

		if i > 0 {
			if _, err := f.WriteString("\n---\n"); err != nil {
				return err
			}
		}

		if _, err := f.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func writeWarnings(path string, warnings []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}(f)

	for _, w := range warnings {
		if _, err := fmt.Fprintln(f, "- "+w); err != nil {
			return err
		}
	}

	return nil
}
