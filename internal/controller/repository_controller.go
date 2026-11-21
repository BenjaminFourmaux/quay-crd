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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"quay-crd/internal/services"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	quayiov1alpha "quay-crd/api/v1alpha"
)

const repositoryFinalizer = "repository.quay.io/finalizer"

// RepositoryReconciler reconciles a Repository object
type RepositoryReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	RepositoryService *services.RepositoryService
}

// +kubebuilder:rbac:groups=quay.io,resources=repositories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=quay.io,resources=repositories/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=quay.io,resources=repositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Repository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Log.Info("[Repository Controller] Reconcile")

	// Get the Repository from kubernetes manifest
	var repository quayiov1alpha.Repository

	err := r.Get(ctx, req.NamespacedName, &repository)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logf.Log.Info("[Repository Controller] Successfully retrieved Repository", "name", repository.Name)

	// Delete
	if !repository.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &repository)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(&repository, repositoryFinalizer) {
		controllerutil.AddFinalizer(&repository, repositoryFinalizer)

		logf.Log.Info("[Repository Controller] Adding Finalizer, comes back to reconcile")

		if err = r.Update(ctx, &repository); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Create Or Update
	return r.reconcileRepository(ctx, &repository)
}

func (r *RepositoryReconciler) reconcileRepository(ctx context.Context, repository *quayiov1alpha.Repository) (ctrl.Result, error) {
	logf.Log.Info("[Repository Controller] Reconcile: CreateOrUpdate", "name", repository.Name)

	needUpdate, err := r.RepositoryService.Reconcile(ctx, repository)
	if err != nil {
		r.setStatus(ctx, repository, err)
		return ctrl.Result{}, err
	}

	// Update if needed
	if needUpdate {
		if err = r.Update(ctx, repository); err != nil {
			r.setStatus(ctx, repository, err)
			return ctrl.Result{}, err
		}
	}

	r.setStatus(ctx, repository, nil)

	return ctrl.Result{}, nil
}

func (r *RepositoryReconciler) reconcileDelete(ctx context.Context, repository *quayiov1alpha.Repository) (ctrl.Result, error) {
	logf.Log.Info("[Repository Controller] Reconcile: Delete", "name", repository.Name)

	if err := r.RepositoryService.Delete(ctx, repository); err != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(repository, repositoryFinalizer)

	if err := r.Update(ctx, repository); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RepositoryReconciler) setStatus(ctx context.Context, repository *quayiov1alpha.Repository, err error) {
	// if err is nil, set status to Ready success
	if err == nil {
		SetCondition(
			&repository.Status.Conditions,
			repository.Generation,
			metav1.ConditionTrue,
			ReasonReconcileSuccess,
			"Reconciliation successful",
		)
	} else {
		SetCondition(
			&repository.Status.Conditions,
			repository.Generation,
			metav1.ConditionFalse,
			GetErrorReason(err),
			err.Error(),
		)
	}

	if statusErr := r.Status().Update(ctx, repository); statusErr != nil {
		logf.Log.Error(statusErr, "Unable to update Repository status", "name", repository.Name)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quayiov1alpha.Repository{}).
		Named("repository").
		Complete(r)
}
