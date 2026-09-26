package services

import (
	"context"
	"errors"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type OrganizationService struct {
	KubeClient    client.Client
	QuayClientMgr *quay.ClientManager
}

/*
NewOrganizationService is the constructor for OrganizationService
*/
func NewOrganizationService(kubeClient client.Client, quayClientMgr *quay.ClientManager) *OrganizationService {
	return &OrganizationService{
		KubeClient:    kubeClient,
		QuayClientMgr: quayClientMgr,
	}
}

func (s *OrganizationService) Reconcile(ctx context.Context, organization *quayiov1alpha.Organization) (bool, error) {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return false, errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Search if the org exists on Quay
	existingOrg, err := s.QuayClientMgr.Get().GetOrganization(organization.Name)
	if err != nil {
		// Create, Org not found in Quay, create it
		if errors.Is(err, quay.ErrNotFound) {
			if err = s.create(ctx, organization); err != nil {
				return false, err
			}
		}
	} else {
		// Chek if Quay state corresponds to desired state, if true return false
		if !needUpdateOrganizationState(organization, existingOrg) {
			logf.Log.Info("[Organization Service] No update needed for organization", "name", organization.Name)
			return false, nil
		} else {
			logf.Log.Info("[Organization Service] Update needed for organization", "name", organization.Name)
		}

		// Update
		if err = s.update(ctx, organization); err != nil {
			return false, err
		}
	}

	// Create or update success, return true for update resource manifest
	return true, nil
}

func (s *OrganizationService) Delete(ctx context.Context, organization *quayiov1alpha.Organization) error {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Delete the organization from Quay
	err := s.QuayClientMgr.Get().DeleteOrganization(organization.Name)

	if err != nil {
		if errors.Is(err, quay.ErrNotFound) {
			// Already deleted in Quay: we assume the delete task is completed
			return nil
		}
		return err
	}
	return nil
}

func (s *OrganizationService) create(ctx context.Context, organization *quayiov1alpha.Organization) error {
	logf.Log.Info("[Organization Service] Creating organization", "name", organization.Name)

	// Create organization Model
	orgToCreate := quay.CreateOrganization{
		Name: organization.Name,
	}

	createdOrg, err := s.QuayClientMgr.Get().CreateOrganization(&orgToCreate)
	if err != nil {
		logf.Log.Error(err, "[Organization Service] Error when creating organization", "name", organization.Name)
		return err
	} else {
		logf.Log.Info("[Organization Service] Successfully created organization", "name", organization.Name)
	}

	// update manifest with Quay's information
	updateOrganizationFromModel(createdOrg, organization)

	// TODO: add owner Team crd

	return nil
}

func (s *OrganizationService) update(ctx context.Context, organization *quayiov1alpha.Organization) error {
	logf.Log.Info("[Organization Service] Updating organization", "name", organization.Name)

	// Prepare Quay Model
	var orgToUpdate quay.UpdateOrganization
	if organization.Spec.Email != nil {
		orgToUpdate.Email = organization.Spec.Email
	}
	if organization.Spec.InvoiceEmail != nil {
		orgToUpdate.InvoiceEmail = organization.Spec.InvoiceEmail
	}
	if organization.Spec.TagExpirationS != nil {
		orgToUpdate.TagExpirationS = organization.Spec.TagExpirationS
	}

	updatedOrg, err := s.QuayClientMgr.Get().UpdateOrganization(organization.Name, &orgToUpdate)
	if err != nil {
		logf.Log.Error(err, "[Organization Service] Error when updating organization", "name", organization.Name)
		return err
	} else {
		logf.Log.Info("[Organization Service] Successfully updated organization", "name", organization.Name)
	}

	// Return the updated org for update the CR manifests
	updateOrganizationFromModel(updatedOrg, organization)

	return nil
}

func updateOrganizationFromModel(quayOrganization quay.Organization, organization *quayiov1alpha.Organization) {
	if quayOrganization.Email != "" {
		organization.Spec.Email = &quayOrganization.Email
	}
	if quayOrganization.InvoiceEmailAddress != nil {
		b := true
		organization.Spec.InvoiceEmail = &b
	}
	if quayOrganization.TagExpirationS != 0 {
		organization.Spec.TagExpirationS = &quayOrganization.TagExpirationS
	}
}

/*
needUpdateOrganizationState Return true if stats are one or more difference, them return false if no update needed
*/
func needUpdateOrganizationState(desiredOrg *quayiov1alpha.Organization, existingOrganization quay.Organization) bool {
	if desiredOrg.Spec.Email != nil && desiredOrg.Spec.Email != existingOrganization.InvoiceEmailAddress {
		return true
	}

	if desiredOrg.Spec.InvoiceEmail != nil && *desiredOrg.Spec.InvoiceEmail != existingOrganization.InvoiceEmail {
		return true
	}

	if desiredOrg.Spec.TagExpirationS != nil && *desiredOrg.Spec.TagExpirationS != existingOrganization.TagExpirationS {
		return true
	}

	return false
}
