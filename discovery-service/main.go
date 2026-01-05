package main

import (
	"context"
	"discovery-service/api"
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/joho/godotenv"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	LogLevel         string
	Port             string
	Host             string
	DebugMode        bool
	VerificationType string
	PublicKey        string
	JwksURL          string
	Alg              string
}

func initConfig() *Config {
	godotenv.Load()

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080" // Default port
	}
	host := os.Getenv("HTTP_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	v_type := os.Getenv("VERIFICATION_TYPE")
	if v_type == "" {
		v_type = "LOCAL"
	}

	debugMode := os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "TRUE"

	return &Config{
		LogLevel:         os.Getenv("LOG_LEVEL"),
		Port:             port,
		Host:             host,
		DebugMode:        debugMode,
		VerificationType: v_type,
		PublicKey:        os.Getenv("PUBLIC_KEY"),
		JwksURL:          os.Getenv("JWKS_URL"),
		Alg:              os.Getenv("JWT_ENCRYPTION"),
	}
}

func createKubeClient(debugMode bool) (*kubernetes.Clientset, error) {
	if debugMode {
		logger.Info("creating kubernetes client in debug mode")
		return api.CreateDebugClient()
	}
	logger.Info("creating in-cluster kubernetes client")
	return api.CreateInClusterClient()
}

func main() {
	config := initConfig()
	InitLogger(config.LogLevel)

	k8sClient, err := createKubeClient(config.DebugMode)
	if err != nil {
		logger.Error("failed to create kubernetes client", "error", err)
		return
	}
	logger.Info("successfully connected to kubernetes")

	ctx := context.Background()

	humaConf := huma.DefaultConfig("geist-discovery-service", "0.0.1")
	humaConf.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	router := chi.NewMux()
	humaAPI := humachi.New(router, humaConf)

	auth, err := api.InitAuth(ctx, config.VerificationType, config.PublicKey, config.Alg, config.JwksURL)

	if err != nil {
		logger.Error("failed to initialize cache", "error", err)
		return
	}

	apiState := api.AppState{
		Auth:         auth,
		HumaInstance: &humaAPI,
		K8s:          k8sClient,
	}

	protected := huma.NewGroup(*apiState.HumaInstance, "/v1")
	protected.UseMiddleware(apiState.RegisterAuthMiddleware)

	huma.Register(protected, huma.Operation{
		Method: http.MethodGet,
		Path:   "/namespace",
	}, apiState.GetNamespace)

	logger.Info("starting server", "host", config.Host, "port", config.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.Host, config.Port), router); err != nil {
		logger.Error("server failed", "error", err)
	}
}
