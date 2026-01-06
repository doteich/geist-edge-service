package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AppState struct {
	Auth         AuthState
	HumaInstance *huma.API
	K8s          K8s
}

type K8s struct {
	Client    client.Client
	Config    *rest.Config
	Namespace string
}

type AuthState struct {
	ValidationType string
	Cache          *jwk.Cache
	JwksURL        string
	Alg            jwa.KeyAlgorithm
	Key            interface{}
}
