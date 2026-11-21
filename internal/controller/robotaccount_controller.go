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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"quay-crd/internal/services"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	quayiov1alpha "quay-crd/api/v1alpha"
)

const robotAccountFinalizer = "robotaccount.quay.io/finalizer"

// RobotAccountReconciler reconciles a RobotAccount object
type RobotAccountReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	RobotAccountService *services.RobotAccountService
}

// +kubebuilder:rbac:groups=quay.io,resources=robotaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=quay.io,resources=robotaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=quay.io,resources=robotaccounts/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RobotAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *RobotAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.Log.Info("[RobotAccount Controller] Reconcile")

	// Get the RobotAccount from kubernetes manifest
	var robotAccount quayiov1alpha.RobotAccount

	err := r.Get(ctx, req.NamespacedName, &robotAccount)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logf.Log.Info("[RobotAccount Controller] Successfully retrieved RobotAccount", "name", robotAccount.Name)

	// Delete
	if !robotAccount.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &robotAccount)
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(&robotAccount, robotAccountFinalizer) {
		controllerutil.AddFinalizer(&robotAccount, robotAccountFinalizer)

		logf.Log.Info("[RobotAccount Controller] Adding Finalizer, comes back to reconcile")

		if err = r.Update(ctx, &robotAccount); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Create Or Update
	return r.reconcileRobotAccount(ctx, &robotAccount)
}

func (r *RobotAccountReconciler) reconcileRobotAccount(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount) (ctrl.Result, error) {
	logf.Log.Info("[RobotAccount Controller] Reconcile: CreateOrUpdate", "name", robotAccount.Name)

	needUpdate, err := r.RobotAccountService.Reconcile(ctx, robotAccount)
	if err != nil {
		r.setStatus(ctx, robotAccount, err)
		return ctrl.Result{}, err
	}

	// Update if needed
	if needUpdate {
		if err = r.Update(ctx, robotAccount); err != nil {
			r.setStatus(ctx, robotAccount, err)
			return ctrl.Result{}, err
		}
	}

	r.setStatus(ctx, robotAccount, nil)

	return ctrl.Result{}, nil
}

func (r *RobotAccountReconciler) reconcileDelete(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount) (ctrl.Result, error) {
	logf.Log.Info("[RobotAccount Controller] Reconcile: Delete", "name", robotAccount.Name)

	if err := r.RobotAccountService.Delete(ctx, robotAccount); err != nil {
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(robotAccount, robotAccountFinalizer)

	if err := r.Update(ctx, robotAccount); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RobotAccountReconciler) setStatus(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount, err error) {
	previousStatus := robotAccount.Status.DeepCopy()
	// if err is nil, set status to Ready success
	if err == nil {
		SetCondition(
			&robotAccount.Status.Conditions,
			robotAccount.Generation,
			metav1.ConditionTrue,
			ReasonReconcileSuccess,
			"Reconciliation successful",
		)
	} else {
		SetCondition(
			&robotAccount.Status.Conditions,
			robotAccount.Generation,
			metav1.ConditionFalse,
			GetErrorReason(err),
			err.Error(),
		)
	}

	if equality.Semantic.DeepEqual(previousStatus, &robotAccount.Status) {
		return
	}
	if statusErr := r.Status().Update(ctx, robotAccount); statusErr != nil {
		logf.Log.Error(statusErr, "Unable to update RobotAccount status", "name", robotAccount.Name)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RobotAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quayiov1alpha.RobotAccount{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, object client.Object) []reconcile.Request {
			annotations := object.GetAnnotations()
			name := annotations[services.RobotAccountNameAnnotation]
			namespace := annotations[services.RobotAccountNamespaceAnnotation]
			if name == "" || namespace == "" || object.GetLabels()[services.RobotAccountUIDLabel] == "" {
				return nil
			}
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}}
		})).
		Named("robotaccount").
		Complete(r)
}
