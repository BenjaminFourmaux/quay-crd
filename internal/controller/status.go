package controller

import (
	"errors"
	apiMeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"quay-crd/internal/services/quay"
)

const (
	ReasonInvalidSpec          = "InvalidSpec"
	ReasonQuayAPIError         = "QuayAPIError"
	ReasonQuayUnauthorized     = "QuayUnauthorized"
	ReasonQuayForbidden        = "QuayForbidden"
	ReasonQuayResourceNotFound = "QuayResourceNotFound"
	ReasonDependencyNotReady   = "DependencyNotReady"
	ReasonReconciliationError  = "ReconciliationError"
	ReasonReconcileSuccess     = "ReconcileSuccess"
)

func SetCondition(
	conditions *[]metav1.Condition,
	generation int64,
	status metav1.ConditionStatus,
	reason string,
	message string,
) {
	apiMeta.SetStatusCondition(
		conditions,
		metav1.Condition{
			Type:               "Ready",
			Status:             status,
			Reason:             reason,
			Message:            message,
			ObservedGeneration: generation,
		})
}

func GetErrorReason(err error) string {
	switch {
	case err == nil:
		return ReasonReconcileSuccess
	case errors.Is(err, quay.ErrNotFound):
		return ReasonQuayResourceNotFound
	default:
		return ReasonReconciliationError
	}
}
