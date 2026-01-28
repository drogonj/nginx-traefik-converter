package convert

import netv1 "k8s.io/api/networking/v1"

type Context struct {
	Ingress     *netv1.Ingress
	IngressName string
	Namespace   string
	Annotations map[string]string
	Result      *Result
}
