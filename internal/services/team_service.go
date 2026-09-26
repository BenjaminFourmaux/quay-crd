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
	existingTeam, err := s.QuayClientMgr.Get().GetTeam(team.Spec.OrganizationRef.Name, team.Spec.Name)
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
		if err = s.update(ctx, team, existingTeam); err != nil {
			return false, err
		}
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
		if errors.Is(err, quay.ErrNotFound) || errors.Is(err, quay.ErrForbidden) {
			// Already deleted in Quay: we assume the delete task is completed
			return nil
		}
		return err
	}
	return nil
}

func (s *TeamService) create(ctx context.Context, team *quayiov1alpha.Team) error {
	logf.Log.Info("[Team Service] Creating team", "org", team.Spec.OrganizationRef.Name, "name", team.Spec.Name)

	// Set OwnerReference to the organization ref
	if err := s.setOwnerReference(ctx, team); err != nil {
		return err
	}

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

func (s *TeamService) update(ctx context.Context, team *quayiov1alpha.Team, existingTeam *quay.Team) error {
	logf.Log.Info("[Team Service] Updating team", "org", team.Spec.OrganizationRef.Name, "name", team.Spec.Name)

	var changed bool = false

	// Prepare Quay Model
	var teamToUpdate quay.UpdateTeam
	if team.Spec.Description != existingTeam.Description {
		teamToUpdate.Description = team.Spec.Description
		changed = true
	}
	if team.Spec.Role != existingTeam.Role {
		teamToUpdate.Role = team.Spec.Role
		changed = true
	}

	if changed {
		err := s.QuayClientMgr.Get().UpdateTeam(team.Spec.OrganizationRef.Name, team.Spec.Name, &teamToUpdate)
		if err != nil {
			return err
		}
	}

	// Retrieve the list of team's member in Quay
	existingMembers, err := s.QuayClientMgr.Get().ListTeamMembers(team.Spec.OrganizationRef.Name, team.Spec.Name)
	if err != nil {
		return err
	}
	existingMembersList := quayMembersToStringList(existingMembers)

	// Update members
	toAdd, toRemove := reconcileList(team.Spec.Members, existingMembersList)

	if len(toRemove) > 0 {
		for _, member := range toRemove {
			if err = s.QuayClientMgr.Get().RemoveTeamMember(team.Spec.OrganizationRef.Name, team.Spec.Name, member); err != nil {
				return err
			}
		}
	}
	if len(toAdd) > 0 {
		for _, member := range toAdd {
			if err = s.QuayClientMgr.Get().AddTeamMember(team.Spec.OrganizationRef.Name, team.Spec.Name, member); err != nil {
				return err
			}
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

func (s *TeamService) setOwnerReference(ctx context.Context, team *quayiov1alpha.Team) error {
	// Retrieve the organization manifest
	var org quayiov1alpha.Organization

	err := s.KubeClient.Get(
		ctx,
		types.NamespacedName{
			Name:      team.Spec.OrganizationRef.Name,
			Namespace: team.Spec.OrganizationRef.Namespace,
		},
		&org,
	)
	if err != nil {
		return err
	}

	// Set ownerReference
	if err = controllerutil.SetControllerReference(
		&org,
		team,
		s.Scheme,
	); err != nil {
		return err
	}
	return nil
}

func formatTeamName(name string, organizationName string) string {
	return organizationName + "-" + name
}

func quayMembersToStringList(members []quay.Member) []string {
	var memberStringList []string
	for _, member := range members {
		memberStringList = append(memberStringList, member.Name)
	}
	return memberStringList
}

/*
reconcileList Desired must be the list from Kubernetes manifest, the current must be the list of Quay
*/
func reconcileList(desired, current []string) (toAdd, toRemove []string) {
	currentSet := make(map[string]struct{}, len(current))

	for _, item := range current {
		currentSet[item] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(desired))

	for _, item := range desired {
		desiredSet[item] = struct{}{}
	}

	// Present in desired but not in current
	for item := range desiredSet {
		if _, exists := currentSet[item]; !exists {
			toAdd = append(toAdd, item)
		}
	}

	// Present in current but not in desired
	for item := range currentSet {
		if _, exists := desiredSet[item]; !exists {
			toRemove = append(toRemove, item)
		}
	}
	return
}

// </editor-fold>
