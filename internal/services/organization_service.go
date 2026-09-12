package services

import (
	"context"
	"errors"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (s *OrganizationService) Reconcile(ctx context.Context, organization *quayiov1alpha.Organization) error {
	// Get the current Quay client from manager
	quayClient := s.QuayClientMgr.Get()
	if quayClient == nil {
		return errors.New("Quay client not initialized - ensure QuayConfig is created and reconciled")
	}

	// Search if the org exists on Quay
	_, err := quayClient.GetOrganization(organization.Name)
	if err != nil {
		if errors.Is(err, quay.ErrNotFound) {
			// Create the organization
			orgToCreate := quay.CreateOrganization{
				Name: organization.Name,
			}

			createdOrg, err := quayClient.CreateOrganization(&orgToCreate)
			if err != nil {
				return err
			}

			// Return the created org for update CR manifest
			updateOrganizationFromModel(createdOrg, organization)

			return nil
		}
	} else {
		// Organization exists, update them
		var orgToUpdate quay.UpdateOrganization

		if organization.Spec.Email != nil {
			orgToUpdate.Email = *organization.Spec.Email
		}
		if organization.Spec.InvoiceEmail != nil {
			orgToUpdate.InvoiceEmail = *organization.Spec.InvoiceEmail
		}
		if organization.Spec.TagExpirationS != nil {
			orgToUpdate.TagExpirationS = *organization.Spec.TagExpirationS
		}

		updatedOrg, err := quayClient.UpdateOrganization(organization.Name, &orgToUpdate)
		if err != nil {
			return err
		}

		// Return the updated org for update the CR manifests
		updateOrganizationFromModel(updatedOrg, organization)

		return nil
	}
	return nil
}

func updateOrganizationFromModel(quayOrganization quay.Organization, organization *quayiov1alpha.Organization) {
	if quayOrganization.Email != "" {
		organization.Spec.Email = &quayOrganization.Email
	}
	if quayOrganization.InvoiceEmailAddress != nil {
		organization.Spec.InvoiceEmail = quayOrganization.InvoiceEmailAddress
	}
	if quayOrganization.TagExpirationS != 0 {
		organization.Spec.TagExpirationS = &quayOrganization.TagExpirationS
	}
}
