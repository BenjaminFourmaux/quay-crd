package services

import (
	"context"
	"errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RepositoryService struct {
	KubeClient    client.Client
	QuayClientMgr *quay.ClientManager
	Scheme        *runtime.Scheme
}

/*
NewRepositoryService is the constructor for RepositoryService
*/
func NewRepositoryService(kubeClient client.Client, quayClientMgr *quay.ClientManager, scheme *runtime.Scheme) *RepositoryService {
	return &RepositoryService{
		KubeClient:    kubeClient,
		QuayClientMgr: quayClientMgr,
		Scheme:        scheme,
	}
}

func (s *RepositoryService) Reconcile(ctx context.Context, repository *quayiov1alpha.Repository) (bool, error) {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return false, errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Search if the repository exists on Quay
	existingRepository, err := s.QuayClientMgr.Get().GetRepository(repository.Spec.Name)
	if err != nil {
		if errors.Is(err, quay.ErrNotFound) {
			// Repository not found, create it
			if err = s.create(ctx, repository); err != nil {
				return false, err
			}
		} else {
			// other errors, raise them
			return false, err
		}
	} else {
		if err = s.update(ctx, repository, &existingRepository); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (s *RepositoryService) Delete(ctx context.Context, repository *quayiov1alpha.Repository) error {
	logf.Log.Info("[Repository Service] Deleting repository", "name", repository.Spec.Name)

	err := s.QuayClientMgr.Get().DeleteRepository(repository.Spec.Name)
	if err != nil {
		logf.Log.Error(err, "[Repository Service] Error when deleting repository", "name", repository.Spec.Name)
		return err
	} else {
		logf.Log.Info("[Repository Service] Repository successfully deleted", "name", repository.Spec.Name)
	}
	
	return nil
}

func (s *RepositoryService) create(ctx context.Context, repository *quayiov1alpha.Repository) error {
	logf.Log.Info("[Repository Service] Creating repository", "name", repository.Spec.Name)

	namespace, repoName := splitRepoFullName(repository)

	// Create the repository Model
	repositoryToCreate := quay.CreateRepository{
		repoName,
		repository.Spec.Visibility,
		namespace,
		repository.Spec.Description,
		"image",
	}

	_, err := s.QuayClientMgr.Get().CreateRepository(&repositoryToCreate)
	if err != nil {
		logf.Log.Error(err, "[Repository Service] Error when creating repository", "name", repository.Spec.Name)
		return err
	} else {
		logf.Log.Info("[Repository Service] Created repository", "name", repository.Spec.Name)
	}

	// TODO: create the repomirroring if condition needed

	// Skip update the manifest, nothing to update I think

	return nil
}

func (s *RepositoryService) update(ctx context.Context, repository *quayiov1alpha.Repository, existingRepository *quay.Repository) error {
	logf.Log.Info("[Repository Service] Updating repository", "name", repository.Spec.Name)

	// Update only description
	if repository.Spec.Description != existingRepository.Description {
		// Prepare Quay Model
		repoToUpdate := quay.UpdateRepository{
			Description: &repository.Spec.Description,
		}

		// Call Quay
		_, err := s.QuayClientMgr.Get().UpdateRepository(repository.Spec.Name, &repoToUpdate)
		if err != nil {
			logf.Log.Info("[Repository Service] Error when updating repository", "name", repository.Spec.Name)
			return err
		} else {
			logf.Log.Info("[Repository Service] Updated repository", "name", repository.Spec.Name)
		}
	}

	// Update Only visibility
	if repository.Spec.Visibility != mapVisibilityToString(existingRepository.IsPublic) {
		// Prepare Quay Model
		repoToUpdate := quay.ChangeRepositoryVisibility{
			Visibility: repository.Spec.Visibility,
		}

		// Call Quay
		err := s.QuayClientMgr.Get().ChangeRepositoryVisibility(repository.Spec.Name, &repoToUpdate)
		if err != nil {
			logf.Log.Info("[Repository Service] Error when updating repository", "name", repository.Spec.Name)
			return err
		} else {
			logf.Log.Info("[Repository Service] Updated repository visibility", "name", repository.Spec.Name, "visibility", repository.Spec.Visibility)
		}
	}

	return nil
}

func splitRepoFullName(repository *quayiov1alpha.Repository) (string, string) {
	parts := strings.SplitN(repository.Spec.Name, "/", 2)
	if len(parts) < 2 {
		return "", repository.Spec.Name
	}

	namespace, repositoryName := parts[0], parts[1]
	return namespace, repositoryName
}

func mapVisibilityToString(isPublic bool) string {
	if isPublic {
		return "public"
	}
	return "private"
}
