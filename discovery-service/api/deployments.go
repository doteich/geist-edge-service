package api

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetDeploymentsInput struct {
	Namespace string `path:"namespace" example:"geist" doc:"The namespace to fetch deployments from"`
}

type DeploymentStatus struct {
	Name          string `json:"name"`
	Replicas      int32  `json:"replicas"`
	ReadyReplicas int32  `json:"readyReplicas"`
	Available     bool   `json:"available"`
}

type GetDeploymentsOutput struct {
	Body struct {
		Redpanda      []DeploymentStatus `json:"redpanda"`
		GeistAPI      DeploymentStatus   `json:"Geist-API"`
		GeistOperator DeploymentStatus   `json:"Geist-Operator"`
	}
}

func (a *AppState) GetDeployments(ctx context.Context, input *GetDeploymentsInput) (*GetDeploymentsOutput, error) {

	resp := &GetDeploymentsOutput{}

	list := &appsv1.DeploymentList{}

	err := a.K8s.Client.List(ctx, list, client.InNamespace(input.Namespace))

	if err != nil {
		return resp, err
	}

	for _, i := range list.Items {

		status := DeploymentStatus{
			Name:      i.Name,
			Available: true,
			Replicas:  1,
		}

		if i.Spec.Replicas != nil {
			status.Replicas = *i.Spec.Replicas
		}
		status.ReadyReplicas = i.Status.ReadyReplicas

		if status.ReadyReplicas < status.Replicas {
			status.Available = false
		}

		isMatched := false
		for _, labelValue := range i.Labels {
			for _, filter := range a.DeploymentFilters.Redpanda {
				if filter != "" && labelValue == filter {
					resp.Body.Redpanda = append(resp.Body.Redpanda, status)
					isMatched = true
					break
				}
			}
			if isMatched {
				break
			}

			if a.DeploymentFilters.GeistAPI != "" && labelValue == a.DeploymentFilters.GeistAPI {
				resp.Body.GeistAPI = status
				isMatched = true
				break
			}

			if a.DeploymentFilters.GeistOperator != "" && labelValue == a.DeploymentFilters.GeistOperator {
				resp.Body.GeistOperator = status
				isMatched = true
				break
			}
		}
	}

	return resp, nil
}
