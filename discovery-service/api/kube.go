package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doteich/geist-edge-service/operator/api/v1alpha"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	commonScheme = runtime.NewScheme()
)

func init() {

	utilruntime.Must(scheme.AddToScheme(commonScheme))
	utilruntime.Must(v1alpha.AddToScheme(commonScheme))
}

func CreateInClusterClient(namespace string) (K8s, error) {

	kube := K8s{}

	config, err := rest.InClusterConfig()
	if err != nil {
		return kube, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	return createClient(config, namespace)
}

func CreateDebugClient(namespace string) (K8s, error) {
	kube := K8s{}

	home, err := os.UserHomeDir()
	if err != nil {
		return kube, fmt.Errorf("failed to get user home directory: %w", err)
	}

	kubeconfigPath := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return kube, fmt.Errorf("failed to build config from %s: %w", kubeconfigPath, err)
	}

	return createClient(config, namespace)

}

func createClient(config *rest.Config, namespace string) (K8s, error) {

	kube := K8s{}

	cl, err := client.New(config, client.Options{Scheme: commonScheme})
	if err != nil {
		return kube, fmt.Errorf("failed to create unified client: %w", err)
	}

	nsList := &corev1.NamespaceList{}
	err = cl.List(context.Background(), nsList, client.Limit(1))
	if err != nil {
		return kube, fmt.Errorf("failed to verify connectivity: %w", err)
	}

	kube.Client = cl
	kube.Config = config
	kube.Namespace = namespace

	return kube, nil
}
