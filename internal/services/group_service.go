package services

import (
	"context"
	"errors"
	"k8s.io/apimachinery/pkg/runtime"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GroupService struct {
	KubeClient    client.Client
	QuayClientMgr *quay.ClientManager
	Scheme        *runtime.Scheme
}

/*
NewGroupService is the constructor for GroupService
*/
func NewGroupService(client client.Client, quayClientMgr *quay.ClientManager, scheme *runtime.Scheme) *GroupService {
	return &GroupService{
		KubeClient:    client,
		QuayClientMgr: quayClientMgr,
		Scheme:        scheme,
	}
}

func (s *GroupService) Reconcile(ctx context.Context, group *quayiov1alpha.Group) (bool, error) {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return false, errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Do nothing on creating

	return false, nil
}

func (s *GroupService) Delete(ctx context.Context, group *quayiov1alpha.Group) error {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	return nil
}
