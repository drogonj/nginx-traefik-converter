package kubernetes

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// certManagerGVR is the GroupVersionResource for cert-manager Certificate CRDs.
var certManagerGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificates",
}

// FindCertificateBySecret lists Certificate resources in the given namespace
// and returns the first one whose spec.secretName matches the provided value.
// Returns (nil, nil) when no matching Certificate is found.
//
// This approach (LIST + filter) is used instead of GET-by-name because
// cert-manager does not guarantee that the Certificate name matches the
// secretName — users may have created Certificates manually with arbitrary
// names.
func (cfg *Config) FindCertificateBySecret(namespace, secretName string) (*unstructured.Unstructured, error) {
	if cfg.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client is not initialised; cannot look up Certificate resources")
	}

	list, err := cfg.dynamicClient.Resource(certManagerGVR).Namespace(namespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		// The cert-manager CRD may not be installed in the cluster.
		// Treat this gracefully: return nil so the caller can fall back
		// to generating a Certificate from annotations.
		if apierrors.IsNotFound(err) {
			if cfg.logger != nil {
				cfg.logger.Warn("cert-manager Certificate CRD not found in cluster; skipping extraction")
			}

			return nil, nil
		}

		return nil, fmt.Errorf("listing cert-manager Certificates in namespace %q: %w", namespace, err)
	}

	for i := range list.Items {
		cert := &list.Items[i]

		specSecret, found, _ := unstructured.NestedString(cert.Object, "spec", "secretName")
		if found && specSecret == secretName {
			return cert, nil
		}
	}

	return nil, nil
}

// SanitizeCertificateForGitOps removes cluster-specific and transient fields
// from a live Certificate resource so it can be safely committed to Git and
// managed declaratively by ArgoCD (or similar).
func SanitizeCertificateForGitOps(cert *unstructured.Unstructured) {
	// Remove server-set metadata that causes drift detection or conflicts.
	unstructured.RemoveNestedField(cert.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(cert.Object, "metadata", "uid")
	unstructured.RemoveNestedField(cert.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(cert.Object, "metadata", "generation")
	unstructured.RemoveNestedField(cert.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(cert.Object, "metadata", "selfLink")

	// Critical: Remove ownerReferences. The Certificate extracted from the
	// cluster points to the Ingress that will be deleted during migration.
	// Without this cleanup cert-manager's garbage collector would delete the
	// Certificate (and its Secret) as soon as the Ingress is removed.
	unstructured.RemoveNestedField(cert.Object, "metadata", "ownerReferences")

	// Remove verbose last-applied-configuration annotation injected by kubectl.
	annotations, _, _ := unstructured.NestedMap(cert.Object, "metadata", "annotations")
	if annotations != nil {
		delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")

		if len(annotations) == 0 {
			unstructured.RemoveNestedField(cert.Object, "metadata", "annotations")
		} else {
			_ = unstructured.SetNestedMap(cert.Object, annotations, "metadata", "annotations")
		}
	}

	// Remove runtime status — it is not declarative.
	unstructured.RemoveNestedField(cert.Object, "status")
}
