/*
Copyright 2024.

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
	"fmt"
	"github.com/go-logr/logr"
	demov1 "hello-k8sOperator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DemoVolumeReconciler reconciles a DemoVolume object
type DemoVolumeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.devops.toolbox,resources=demovolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.devops.toolbox,resources=demovolumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.devops.toolbox,resources=demovolumes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DemoVolume object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *DemoVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered Reconcile", "req", req)

	volume := &demov1.DemoVolume{}

	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, volume)

	if volume.Spec.Name != volume.Status.Name {
		volume.Status.Name = volume.Spec.Name
		r.Status().Update(ctx, volume)
	}

	r.reconvileSVC(ctx, volume, l)
	r.ReconcileDeployment(ctx, volume, l, "dockerbogo/docker-nginx-hello-world")

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

func (r *DemoVolumeReconciler) reconvileSVC(ctx context.Context, volume *demov1.DemoVolume, l logr.Logger) error {

	svc := &v1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: volume.Name, Namespace: volume.Namespace}, svc)

	if err == nil {
		l.Info("SVC found!")
		return nil
	}

	// Creates a service
	svc = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: volume.Namespace,
			Name:      volume.Name,
			Labels: map[string]string{
				"app": "mycoolapp",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Port: 80,
					Name: "http", // Optional: Define a name for the port
				},
			},
			Selector: map[string]string{"app": "myapp"},
		},
	}

	//Creates

	l.Info("trying to create service!")

	return r.Create(ctx, svc)
}

func (r *DemoVolumeReconciler) ReconcileDeployment(ctx context.Context, volume *demov1.DemoVolume, l logr.Logger, imageName string) {

	replicas := int32(2) // Number of pod replicas

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: volume.Namespace,
			Name:      volume.Name,
			Labels: map[string]string{
				"app": "mycoolapp",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "myapp"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "myapp"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "my-container",
							Image: imageName,
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	} //,
	//}

	l.Info("Trying to create deployment!")

	err := r.Create(ctx, deployment)

	if err != nil {
		fmt.Println("Error creating deployment:", err)
		return
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *DemoVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.DemoVolume{}).
		Complete(r)
}
