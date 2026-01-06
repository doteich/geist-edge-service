package api

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

type GetNamespaceOutput struct {
	Body struct {
		Namespace string `json:"namespace"`
	}
}

func (a *AppState) GetDeployment(ctx context.Context, input *struct{}) (*DeploymentOutput, error) {
	// n, _ := a.K8s.CoreV1().Namespaces().Get(ctx, "geist", v1.GetOptions{})
	// fmt.Println(n.Name)
	resp := &DeploymentOutput{}

	d := &appsv1.Deployment{}

	a.K8s.Client.Get(ctx, types.NamespacedName{Namespace: a.K8s.Namespace}, d)

	fmt.Println(d)

	return resp, nil
}
