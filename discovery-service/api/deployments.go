package api

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetDeploymentsInput struct {
	Namespace string `path:"namespace" example:"geist" doc:"The namespace to fetch deployments from"`
}

type GetDeploymentsOutput struct {
	Body struct {
		Deployments []appsv1.Deployment `json:"deployments"`
	}
}

func (a *AppState) GetDeployments(ctx context.Context, input *GetDeploymentsInput) (*GetDeploymentsOutput, error) {

	resp := &GetDeploymentsOutput{}

	list := &appsv1.DeploymentList{}

	err := a.K8s.Client.List(ctx, list, client.InNamespace(input.Namespace))

	if err != nil {
		return resp, err
	}

	redacted := make([]appsv1.Deployment, 0)

	for _, i := range list.Items {
		for k := range i.Annotations {
			delete(i.Annotations, k)
		}
		i.ManagedFields = make([]v1.ManagedFieldsEntry, 0)
		redacted = append(redacted, i)
	}

	resp.Body.Deployments = redacted

	return resp, nil
}
