package convert

import (
	"fmt"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_bodySize(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var scheme = runtime.NewScheme()

		_ = traefik.AddToScheme(scheme)

		fmt.Println(traefik.SchemeGroupVersion.String())
		fmt.Println(traefik.Kind("middleware"))
		fmt.Println(traefik.Middleware.GetName())
	})
}
