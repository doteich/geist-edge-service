/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"

	"github.com/doteich/geist-edge-service/operator/api/v1alpha"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GeistConnectorReconciler reconciles a GeistConnector object
type GeistConnectorReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	OperatorNamespace string
}

//+kubebuilder:rbac:groups=config.geist-iot.com,resources=geistconnectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.geist-iot.com,resources=geistconnectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.geist-iot.com,resources=geistconnectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GeistConnector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *GeistConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. Fetch the GeistConnector instance
	var geistConnector v1alpha.GeistConnector

	if err := r.Get(ctx, req.NamespacedName, &geistConnector); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("GeistConnector resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get GeistConnector")
		return ctrl.Result{}, err
	}

	// 2. Reconcile the ConfigMap
	cm, err := r.desiredConfigMap(&geistConnector)
	if err != nil {
		log.Error(err, "Failed to define desired ConfigMap")
		return ctrl.Result{}, err
	}

	// Check if ConfigMap already exists
	foundCM := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
		if err = r.Create(ctx, cm); err != nil {
			log.Error(err, "Failed to create new ConfigMap")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	secret, err := r.desiredSecret(&geistConnector)
	if err != nil {
		log.Error(err, "Failed to define desired ConfigMap")
		return ctrl.Result{}, err
	}

	if !geistConnector.Spec.ConnectorSpec.OPCUA.Connection.Certificate.AutoCreate && !geistConnector.Spec.ConnectorSpec.OPCUA.Connection.Certificate.ExternalCertificate {

		foundSecret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, foundSecret)

		if err != nil && apierrors.IsNotFound(err) {
			log.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
			if err = r.Create(ctx, secret); err != nil {
				log.Error(err, "Failed to create new Secret")
				return ctrl.Result{}, err
			}
		} else if err != nil {
			log.Error(err, "Failed to get Secret")
			return ctrl.Result{}, err
		}

	}

	// 3. Reconcile the Deployment
	deployment, err := r.desiredDeployment(&geistConnector)
	if err != nil {
		log.Error(err, "Failed to define desired Deployment")
		return ctrl.Result{}, err
	}

	// Check if Deployment already exists
	foundDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		if err = r.Create(ctx, deployment); err != nil {
			log.Error(err, "Failed to create new Deployment")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Note: For a production controller, you'd add logic here to update the Deployment if it differs.

	// 4. Update Status (Example)
	// You might want to update the status to reflect that resources are created
	// geistConnector.Status.Conditions = ...
	// if err := r.Status().Update(ctx, &geistConnector); err != nil {
	// 	log.Error(err, "Failed to update GeistConnector status")
	// 	return ctrl.Result{}, err
	// }

	log.Info("Reconciliation successful")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GeistConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha.GeistConnector{}).
		Named("geistconnector").
		Complete(r)
}

// desiredConfigMap defines the desired ConfigMap object for a GeistConnector
func (r *GeistConnectorReconciler) desiredConfigMap(gc *v1alpha.GeistConnector) (*corev1.ConfigMap, error) {

	bArr, err := json.Marshal(gc.Spec.ConnectorSpec)

	if err != nil {
		return nil, err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gc.Name + "-config",
			Namespace: r.OperatorNamespace,                 // Use namespace from spec
			Labels:    gc.Spec.DeploymentSpec.CustomLabels, // Use labels from spec
		},
		Data: map[string]string{
			"config.yaml": string(bArr),
		},
	}

	// Set GeistConnector instance as the owner and controller
	if err := ctrl.SetControllerReference(gc, cm, r.Scheme); err != nil {
		return nil, err
	}
	return cm, nil
}

// desiredDeployment defines the desired Deployment object for a GeistConnector
func (r *GeistConnectorReconciler) desiredDeployment(gc *v1alpha.GeistConnector) (*appsv1.Deployment, error) {
	labels := gc.Spec.DeploymentSpec.CustomLabels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app.kubernetes.io/name"] = "geist-connector"
	labels["app.kubernetes.io/instance"] = gc.Name

	configMapName := gc.Name + "-config"

	imageName := gc.Spec.DeploymentSpec.ImageRepo + ":" + gc.Spec.DeploymentSpec.ImageVersion

	var pp corev1.PullPolicy

	switch gc.Spec.DeploymentSpec.PullPolicy {
	case "Never":
		pp = corev1.PullAlways
	case "IfNotPresent":
		pp = corev1.PullIfNotPresent
	default:
		pp = corev1.PullAlways
	}

	var replicas int32 = 1

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        gc.Name + "-deployment",
			Namespace:   r.OperatorNamespace,
			Labels:      labels,
			Annotations: gc.Spec.DeploymentSpec.CustomAnnotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas), // Default to one replica
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           imageName,
						Name:            "geist-connector",
						ImagePullPolicy: pp,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080, // Example port
							Name:          "http",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config-volume",
							MountPath: "/etc/config", // Path inside the container
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "config-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMapName,
								},
							},
						},
					}},
				},
			},
		},
	}

	secretName := gc.Name + "-secret"
	if !gc.Spec.ConnectorSpec.OPCUA.Connection.Certificate.AutoCreate && !gc.Spec.ConnectorSpec.OPCUA.Connection.Certificate.ExternalCertificate {

		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: "certificates",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "certificates",
			MountPath: "/app/certs", // Path inside the container
		})
	}

	// Set GeistConnector instance as the owner and controller
	if err := ctrl.SetControllerReference(gc, deployment, r.Scheme); err != nil {
		return nil, err
	}

	deployment.GetAnnotations()
	return deployment, nil
}

func (r *GeistConnectorReconciler) desiredSecret(gc *v1alpha.GeistConnector) (*corev1.Secret, error) {

	labels := gc.Spec.DeploymentSpec.CustomLabels
	if labels == nil {
		labels = make(map[string]string)
	}
	data := make(map[string][]byte, 0)

	data["cert.pem"] = []byte(gc.Spec.ConnectorSpec.OPCUA.Connection.Certificate.Certificate)
	data["key.pem"] = []byte(gc.Spec.ConnectorSpec.OPCUA.Connection.Certificate.Key)

	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        gc.Name + "-secret",
			Namespace:   r.OperatorNamespace,
			Labels:      labels,
			Annotations: gc.Spec.DeploymentSpec.CustomAnnotations,
		},
		Data: data,
	}

	if err := ctrl.SetControllerReference(gc, &sec, r.Scheme); err != nil {
		return nil, err
	}

	return &sec, nil

}
