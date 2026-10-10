package services

import (
	"context"
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

	// Retrieve Teams that reference this group
	var teams quayiov1alpha.TeamList

	if err := s.KubeClient.List(
		ctx,
		&teams,
		client.InNamespace(group.Namespace),
	); err != nil {
		return err
	}

	// Browse all teams and filter on only teams that reference this group

	for _, team := range teams.Items {
		referencesGroup := false

		for _, groupRef := range team.Spec.Groups {
			if groupRef.Name == group.Name {
				referencesGroup = true
				break
			}
		}

		if !referencesGroup {
			continue
		}

		// Get the list of all members (direct referenced in team member and group) without considering the deleting group's members
		membersAfter, err := s.listMembersWithoutGroup(ctx, &team, group)
		if err != nil {
			return err
		}

		// Retrieve the list of current team on quay
		existingMembers, err := s.QuayClientMgr.Get().ListTeamMembers(team.Spec.OrganizationRef.Name, team.Spec.Name)
		if err != nil {
			return err
		}
		currentMembers := quayMembersToStringList(existingMembers)

		_, toRemove := reconcileList(uniqueStrings(membersAfter), currentMembers)

		if len(toRemove) > 0 {
			for _, member := range toRemove {
				if err = s.QuayClientMgr.Get().RemoveTeamMember(team.Spec.OrganizationRef.Name, team.Spec.Name, member); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *GroupService) listMembersWithoutGroup(ctx context.Context, team *quayiov1alpha.Team, groupToRemove *quayiov1alpha.Group) ([]string, error) {
	var members []string

	// add members from direct team reference
	members = append(members, team.Spec.Members...)

	// Add members from group reference excepted the group we want to delete
	for _, group := range team.Spec.Groups {
		// Retrieve the group manifest from k8s
		var grp quayiov1alpha.Group
		err := s.KubeClient.Get(
			ctx,
			types.NamespacedName{
				Name:      group.Name,
				Namespace: team.Namespace,
			},
			&grp,
		)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logf.Log.Info("[Group Service] Group resource not found, skipping", "group", group.Name, "namespace", team.Namespace)
				continue
			}
			return nil, err
		}

		if grp.Namespace != team.Namespace {
			members = append(members, grp.Spec.Members...)
		} else {
			// Is the group we want to delete
			continue
		}
	}
	return members, nil
}
