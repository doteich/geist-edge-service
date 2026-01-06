package api

import (
	"context"

	"github.com/doteich/geist-edge-service/operator/api/v1alpha"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetCRDsOutput struct {
	Body struct {
		CRDs []v1alpha.GeistConnector `json:"crds"`
	}
}

func (a *AppState) GetCRDs(ctx context.Context, input *struct{}) (*GetCRDsOutput, error) {

	resp := &GetCRDsOutput{}

	list := &v1alpha.GeistConnectorList{}

	err := a.K8s.Client.List(ctx, list, client.InNamespace("geist"))

	if err != nil {
		return resp, err
	}

	redacted := make([]v1alpha.GeistConnector, 0)

	for _, i := range list.Items {
		for k := range i.Annotations {
			delete(i.Annotations, k)
		}
		if i.Spec.ConnectorSpec.OPCUA.Connection.Authentication.Credentials.Password != "" {
			i.Spec.ConnectorSpec.OPCUA.Connection.Authentication.Credentials.Password = "[REDACTED]"
		}
		if i.Spec.ConnectorSpec.OPCUA.Connection.Certificate.Certificate != "" {
			i.Spec.ConnectorSpec.OPCUA.Connection.Certificate.Certificate = "[REDACTED]"
		}
		if i.Spec.ConnectorSpec.OPCUA.Connection.Certificate.Key != "" {
			i.Spec.ConnectorSpec.OPCUA.Connection.Certificate.Key = "[REDACTED]"
		}

		i.ManagedFields = make([]v1.ManagedFieldsEntry, 0)
		redacted = append(redacted, i)
	}

	resp.Body.CRDs = redacted

	return resp, nil
}
