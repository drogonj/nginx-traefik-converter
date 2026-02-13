package render

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const dirPermission = 0o755

// WriteYAML writes the translated inputs to respective files.
func WriteYAML(res configs.Result, outDir string) error {
	if err := os.MkdirAll(outDir, dirPermission); err != nil {
		return err
	}

	if err := writeObjects(
		filepath.Join(outDir, "middlewares.yaml"),
		toClientObjects(res.Middlewares),
	); err != nil {
		return err
	}

	if err := writeObjects(
		filepath.Join(outDir, "ingressroutes.yaml"),
		toClientObjects(res.IngressRoutes),
	); err != nil {
		return err
	}

	if err := writeObjects(
		filepath.Join(outDir, "tlsoptions.yaml"),
		toClientObjects(res.TLSOptions),
	); err != nil {
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

func toClientObjects[T client.Object](in []T) []client.Object {
	out := make([]client.Object, 0, len(in))
	for _, o := range in {
		out = append(out, o)
	}

	return out
}

func writeObjects(path string, objs []client.Object) error {
	if len(objs) == 0 {
		return nil
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	for index, obj := range objs {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}

		if index > 0 {
			if _, err = file.WriteString("\n---\n"); err != nil {
				return err
			}
		}

		if _, err = file.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func writeWarnings(path string, warnings []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	for _, w := range warnings {
		if _, err = fmt.Fprintln(file, "- "+w); err != nil {
			return err
		}
	}

	return nil
}
