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
	"quay-crd/internal/services"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	quayiov1alpha "quay-crd/api/v1alpha"
)

const groupFinalizer = "group.quay.io/finalizer"

// GroupReconciler reconciles a Group object
type GroupReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	GroupService *services.GroupService
}

// +kubebuilder:rbac:groups=quay.io,resources=groups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=quay.io,resources=groups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=quay.io,resources=groups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Group object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Log.Info("[Group Controller] Reconcile")

	// Get the Group from Kubernetes manifest
	var group quayiov1alpha.Group

	err := r.Get(ctx, req.NamespacedName, &group)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logf.Log.Info("[Group Controller] Successfully retrieved Group", "name", group.Name)

	// Delete
	if !group.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &group)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(&group, groupFinalizer) {
		controllerutil.AddFinalizer(&group, groupFinalizer)

		logf.Log.Info("[Group Controller] Adding Finalizer, comes back to reconcile")

		if err = r.Update(ctx, &group); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Create Or Update
	return r.reconcileGroup(ctx, &group)
}

func (r *GroupReconciler) reconcileGroup(ctx context.Context, group *quayiov1alpha.Group) (ctrl.Result, error) {
	logf.Log.Info("[Group Controller] Reconcile: CreateOrUpdate", "name", group.Name)

	needUpdate, err := r.GroupService.Reconcile(ctx, group)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update if needed
	if needUpdate {
		if err = r.Update(ctx, group); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *GroupReconciler) reconcileDelete(ctx context.Context, group *quayiov1alpha.Group) (ctrl.Result, error) {
	logf.Log.Info("[Group Controller] Reconcile: Delete", "name", group.Name)

	if err := r.GroupService.Delete(ctx, group); err != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(group, groupFinalizer)

	if err := r.Update(ctx, group); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quayiov1alpha.Group{}).
		Named("group").
		Complete(r)
}
