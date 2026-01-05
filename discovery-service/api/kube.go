package api

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateInClusterClient() (*kubernetes.Clientset, error) {
	// 1. Get the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// 2. Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	ctx := context.TODO()

	_, err = clientset.CoreV1().Namespaces().Get(ctx, "default", v1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to ping kubernetes server: %w", err)
	}

	return clientset, nil
}

// CreateDebugClient connects using a local kubeconfig file (e.g., ~/.kube/config).
// This is useful for running the app on your local Windows machine during development.
func CreateDebugClient() (*kubernetes.Clientset, error) {
	// 1. Find the user's home directory to locate .kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Define the path to the kubeconfig file
	// On Windows, this resolves to C:\Users\YourName\.kube\config
	kubeconfigPath := filepath.Join(home, ".kube", "config")

	// Allow overriding via a flag if needed (optional but good practice)
	// You can remove these 3 lines if you want to hardcode the path strictly
	var kubeconfig *string
	if flag.Lookup("kubeconfig") == nil {
		kubeconfig = flag.String("kubeconfig", kubeconfigPath, "(optional) absolute path to the kubeconfig file")
		flag.Parse()
	} else {
		// If flag is already defined in main, just use the path string
		s := kubeconfigPath
		kubeconfig = &s
	}

	// 2. Build configuration from the config file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from %s: %w", *kubeconfig, err)
	}

	// 3. Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	ctx := context.TODO()

	_, err = clientset.CoreV1().Namespaces().Get(ctx, "default", v1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to ping kubernetes server: %w", err)
	}

	return clientset, nil
}
