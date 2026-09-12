/*
Copyright 2026.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	quayiov1alpha1 "quay-crd/api/v1alpha"
	"quay-crd/internal/services"
	"quay-crd/internal/services/quay"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// OrganizationReconciler reconciles a Organization object
type OrganizationReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	OrganizationService *services.OrganizationService
}

// +kubebuilder:rbac:groups=quay.io,resources=organizations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=quay.io,resources=organizations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=quay.io,resources=organizations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Organization object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here
	logf.Log.Info("Coucou")

	// 1. Get the Organization from Kubernetes manifest
	var org quayiov1alpha1.Organization

	err := r.Get(ctx, req.NamespacedName, &org)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logf.Log.Info("Successfully retrieved Organization", "name", org.Name)

	// 2. Call service
	err = r.OrganizationService.Reconcile(ctx, &org)
	if err != nil {
		return ctrl.Result{}, err
	} else {
		// Update CR manifest
		if err = r.Update(ctx, &org); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) setNotReady(
	ctx context.Context,
	org *quayiov1alpha1.Organization,
	reason string,
	message string,
) error {
	meta.SetStatusCondition(&org.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})

	err := r.Status().Update(ctx, org)
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func (r *OrganizationReconciler) setReady(
	ctx context.Context,
	org *quayiov1alpha1.Organization,
	reason string,
	message string,
) error {
	meta.SetStatusCondition(&org.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})

	err := r.Status().Update(ctx, org)
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func conditionFromQuayError(err error) (string, string) {
	apiErr, ok := err.(*quay.APIError)
	if !ok {
		return "QuayRequestFailed", err.Error()
	}

	reason := "QuayInvalidRequest"
	if apiErr.ErrorType != "" {
		reason = "Quay" + toReasonPart(apiErr.ErrorType)
	}

	message := apiErr.Detail
	if message == "" {
		message = apiErr.ErrorMessage
	}
	if message == "" {
		message = apiErr.Error()
	}

	return reason, message
}

func toReasonPart(input string) string {
	result := make([]rune, 0, len(input))
	capitalizeNext := true
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			if capitalizeNext && r >= 'a' && r <= 'z' {
				r = r - 'a' + 'A'
			}
			result = append(result, r)
			capitalizeNext = false
			continue
		}
		capitalizeNext = true
	}
	if len(result) == 0 {
		return "RequestFailed"
	}
	return string(result)
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quayiov1alpha1.Organization{}).
		Named("organization").
		Complete(r)
}
