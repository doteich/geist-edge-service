package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"k8s.io/client-go/kubernetes"
)

type AppState struct {
	Auth         AuthState
	HumaInstance *huma.API
	K8s          *kubernetes.Clientset
}

type AuthState struct {
	ValidationType string
	Cache          *jwk.Cache
	JwksURL        string
	Alg            jwa.KeyAlgorithm
	Key            interface{}
}
