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

const teamFinalizer = "team.quay.io/finalizer"

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TeamService *services.TeamService
}

// +kubebuilder:rbac:groups=quay.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=quay.io,resources=teams/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=quay.io,resources=teams/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Team object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Log.Info("[Team Controller] Reconcile")

	// Get the Team from kubernetes manifest
	var team quayiov1alpha.Team

	err := r.Get(ctx, req.NamespacedName, &team)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logf.Log.Info("[Team Controller] Successfully retrieved Team", "name", team.Name)

	// Delete
	if !team.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &team)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(&team, teamFinalizer) {
		controllerutil.AddFinalizer(&team, teamFinalizer)

		logf.Log.Info("[Team Controller] Adding Finalizer, comes back to reconcile")

		if err = r.Update(ctx, &team); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Create Or Update
	return r.reconcileTeam(ctx, &team)
}

func (r *TeamReconciler) reconcileTeam(ctx context.Context, team *quayiov1alpha.Team) (ctrl.Result, error) {
	logf.Log.Info("[Team Controller] Reconcile: CreateOrUpdate", "name", team.Name)

	needUpdate, err := r.TeamService.Reconcile(ctx, team)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update if needed
	if needUpdate {
		if err = r.Update(ctx, team); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *TeamReconciler) reconcileDelete(ctx context.Context, team *quayiov1alpha.Team) (ctrl.Result, error) {
	logf.Log.Info("[Team Controller] Reconcile: Delete", "name", team.Name)

	if err := r.TeamService.Delete(ctx, team); err != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(team, teamFinalizer)

	if err := r.Update(ctx, team); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quayiov1alpha.Team{}).
		Named("team").
		Complete(r)
}
