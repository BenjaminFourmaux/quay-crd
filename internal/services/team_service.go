package services

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type TeamService struct {
	KubeClient    client.Client
	QuayClientMgr *quay.ClientManager
	Scheme        *runtime.Scheme
}

/*
NewTeamService is the constructor for TeamService
*/
func NewTeamService(kubeClient client.Client, quayClientMgr *quay.ClientManager, scheme *runtime.Scheme) *TeamService {
	return &TeamService{
		KubeClient:    kubeClient,
		QuayClientMgr: quayClientMgr,
		Scheme:        scheme,
	}
}

func (s *TeamService) Reconcile(ctx context.Context, team *quayiov1alpha.Team) (bool, error) {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return false, errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Search if the team exists on Quay
	_, err := s.QuayClientMgr.Get().GetTeam(team.Spec.OrganizationRef.Name, team.Spec.Name)
	if err != nil {
		if errors.Is(err, quay.ErrNotFound) { // Org not found
			return false, err
		} else if err.Error() == "Team not found" {
			// Team not found, create it
			if err = s.create(ctx, team); err != nil {
				return false, err
			}
		} else {
			// other errors, raise them
			return false, err
		}
	} else {
		// TODO: update
	}

	return true, nil
}

func (s *TeamService) Delete(ctx context.Context, team *quayiov1alpha.Team) error {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Check if the team is the owners team than return an error (owners team must be not removed)
	if team.Spec.Name == "owners" && team.Spec.Role == "admin" {
		logf.Log.Info("[Team Service] Try to delete the 'owners' team. Checking if organization exists")

		orgExist, _ := s.checkIfOrganizationCrdExists(ctx, team.Spec.OrganizationRef.Name, team.Spec.OrganizationRef.Namespace)

		// If the 'owners' organization exists on Kubernetes, we cannot remove this team
		if orgExist {
			logf.Log.Info("[Team Service] Organization exists. We cannot delete this Team")
			return errors.New("Deleting Team 'owners' is not allowed")
		} else {
			logf.Log.Info("[Team Service] Organization does not exist. We can delete this Team")
			return nil
		}
	}

	// Delete the team from Quay
	err := s.QuayClientMgr.Get().DeleteTeam(team.Spec.OrganizationRef.Name, team.Spec.Name)

	if err != nil {
		if errors.Is(err, quay.ErrNotFound) {
			// Already deleted in Quay: we assume the delete task is completed
			return nil
		}
		return err
	}
	return nil
}

func (s *TeamService) create(ctx context.Context, team *quayiov1alpha.Team) error {
	logf.Log.Info("[Team Service] Creating team", "org", team.Spec.OrganizationRef.Name, "name", team.Spec.Name)

	// TODO: set ownerReference

	// Create the team Model
	teamToCreate := quay.UpdateTeam{
		Description: team.Spec.Description,
		Role:        team.Spec.Role,
	}

	err := s.QuayClientMgr.Get().UpdateTeam(team.Spec.OrganizationRef.Name, team.Spec.Name, &teamToCreate)
	if err != nil {
		logf.Log.Error(err, "[Team Service] Error when creating team", "org", team.Spec.OrganizationRef.Name, "name", team.Spec.Name)
		return err
	} else {
		logf.Log.Info("[Team Service] Successfully created team", "org", team.Spec.OrganizationRef.Name, "name", team.Spec.Name)
	}

	// Add specified team members
	// TODO: getting members from group crd if specified in this manifest
	for _, member := range team.Spec.Members {
		logf.Log.Info("[Team Service] Adding member", "org", team.Spec.OrganizationRef.Name, "team", team.Spec.Name, "member", member)

		err = s.QuayClientMgr.Get().AddTeamMember(team.Spec.OrganizationRef.Name, team.Spec.Name, member)
		if err != nil {
			logf.Log.Error(err, "[Team Service] Error when adding team member", "org", team.Spec.OrganizationRef.Name, "team", team.Spec.Name, "member", member)
			return err
		} else {
			logf.Log.Info("[Team Service] Member successfully added", "org", team.Spec.OrganizationRef.Name, "team", team.Spec.Name, "member", member)
		}
	}

	return nil
}

func (s *TeamService) CreateOwnersTeam(ctx context.Context, organization *quayiov1alpha.Organization, organizationNamespace string) error {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Call Quay to retrieve owners team member (the current user)
	members, err := s.QuayClientMgr.Get().ListTeamMembers(organization.Name, "owners")
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
			Name:      formatTeamName("owners", organization.Name),
			Namespace: organizationNamespace,
		},
		Spec: quayiov1alpha.TeamSpec{
			Name:    "owners",
			Members: membersStr,
			Role:    "admin",
			OrganizationRef: corev1.ObjectReference{
				Name:       organization.Name,
				Kind:       "Organization",
				APIVersion: "quay.io/v1alpha",
				Namespace:  organizationNamespace,
			},
		},
	}

	// Set ownerReference
	if err = controllerutil.SetControllerReference(
		organization,
		teamCrd,
		s.Scheme,
	); err != nil {
		return err
	}

	// Deploy the crd on Kubernetes
	err = s.KubeClient.Create(ctx, teamCrd)
	if err != nil {
		return err
	}

	return nil
}

func (s *TeamService) DeleteAssociateTeams(ctx context.Context, organizationName string, organizationNamespace string) error {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}
	return nil
}

// <editor-fold desc="Private methods">

func (s *TeamService) checkIfOrganizationCrdExists(ctx context.Context, organizationCrdName string, namespace string) (bool, error) {
	var org quayiov1alpha.Organization

	err := s.KubeClient.Get(
		ctx,
		types.NamespacedName{
			Name:      organizationCrdName,
			Namespace: namespace,
		},
		&org,
	)

	if err != nil {
		return false, err
	}
	return true, nil
}

func formatTeamName(name string, organizationName string) string {
	return organizationName + "-" + name
}

// </editor-fold>
