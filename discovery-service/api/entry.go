package api

import (
	"context"
	"fmt"

	"k8s.io/client-go/discovery"
)

type ConnectivityCheckOutput struct {
	Body struct {
		KubeVersion string `json:"kubernetesVersion"`
		Namespace   string `json:"namespace"`
	}
}

func (a *AppState) ConnectivityCheck(ctx context.Context, input *struct{}) (*ConnectivityCheckOutput, error) {

	resp := ConnectivityCheckOutput{}

	dc, err := discovery.NewDiscoveryClientForConfig(a.K8s.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// 2. Fetch version info
	version, err := dc.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s version: %w", err)
	}

	resp.Body.KubeVersion = fmt.Sprintf("%s.%s", version.Major, version.Minor)
	resp.Body.Namespace = a.K8s.Namespace

	return &resp, nil

}
