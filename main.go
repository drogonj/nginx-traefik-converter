package main

import (
	"log"
	"os"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/convert"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/ingress"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/render"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: nginx2traefik <ingress.yaml>")
	}

	ing, err := ingress.Load(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	res := convert.Result{}
	ctx := convert.Context{
		Ingress:     ing,
		IngressName: ing.Name,
		Namespace:   ing.Namespace,
		Annotations: ing.Annotations,
		Result:      &res,
	}

	convert.Run(ctx)

	if err := render.WriteYAML(res, "./out"); err != nil {
		log.Fatal(err)
	}
}
