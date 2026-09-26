package services

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TeamService struct {
	KubeClient    client.Client
	QuayClientMgr *quay.ClientManager
}

/*
NewTeamService is the constructor for TeamService
*/
func NewTeamService(kubeClient client.Client, quayClientMgr *quay.ClientManager) *TeamService {
	return &TeamService{
		KubeClient:    kubeClient,
		QuayClientMgr: quayClientMgr,
	}
}

func (s *TeamService) Reconcile(ctx context.Context, team *quayiov1alpha.Team) (bool, error) {
	return true, nil
}

func (s *TeamService) CreateOwnersTeam(ctx context.Context, organizationName string, organizationNamespace string) error {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Call Quay to retrieve owners team member (the current user)
	members, err := s.QuayClientMgr.Get().ListTeamMembers(organizationName, "owners")
	if err != nil {
		return err
	}

	var membersStr []string
	for _, member := range members {
		membersStr = append(membersStr, member.Name)
	}

	// Create the team crd
	teamCrd := &quayiov1alpha.Team{
		ObjectMeta: metav1.ObjectMeta{
			Name:      formatTeamName("owners", organizationName),
			Namespace: organizationNamespace,
		},
		Spec: quayiov1alpha.TeamSpec{
			Name:    "owners",
			Members: membersStr,
			Role:    "admin",
			OrganizationRef: corev1.ObjectReference{
				Name:       organizationName,
				Kind:       "Organization",
				APIVersion: "quay.io/v1alpha",
				Namespace:  organizationNamespace,
			},
		},
	}

	// Deploy the crd on Kubernetes
	err = s.KubeClient.Create(ctx, teamCrd)
	if err != nil {
		return err
	}

	return nil
}

func formatTeamName(name string, organizationName string) string {
	return organizationName + "-" + name
}
