package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type State struct {
	K8s *kubernetes.Clientset
}

func main() {
	os.Setenv("DEBUG", "true")

	l_lvl := os.Getenv("LOG_LEVEL")
	d_env := os.Getenv("DEBUG")
	port := os.Getenv("HTTP_PORT")
	host := "0.0.0.0"

	d_mode := false

	if d_env == "true" {
		d_mode = true
		host = "127.0.0.1"
	}

	InitLogger(l_lvl)

	var kube_err error
	var k8s_client *kubernetes.Clientset

	if d_mode {
		k8s_client, kube_err = CreateDebugClient()
	} else {
		k8s_client, kube_err = CreateInClusterClient()
	}

	if kube_err != nil {
		logger.Error(kube_err.Error())
		return
	}

	state := &State{
		K8s: k8s_client,
	}

	logger.Info("successfully connected to kubernetes")

	router := chi.NewMux()

	router.Use()

	humachi.New(router, huma.DefaultConfig("geist-discovery-service", "0.0.1"))

	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), router); err != nil {
		logger.Error(err.Error())
		return
	}

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(30*time.Second), )
	tokenAuth.ValidateOptions()
}

func (s *State) GetNamespace() {
	ctx := context.Background()

	n, _ := s.K8s.CoreV1().Namespaces().Get(ctx, "geist", v1.GetOptions{})
	fmt.Println(n.Name)
}
