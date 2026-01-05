package api

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetNamespaceOutput struct {
	Body struct {
		Namespace string `json:"namespace"`
	}
}

func (a *AppState) GetNamespace(ctx context.Context, input *struct{}) (*GetNamespaceOutput, error) {
	n, _ := a.K8s.CoreV1().Namespaces().Get(ctx, "geist", v1.GetOptions{})
	fmt.Println(n.Name)
	resp := &GetNamespaceOutput{}
	resp.Body.Namespace = n.Name

	return resp, nil
}
