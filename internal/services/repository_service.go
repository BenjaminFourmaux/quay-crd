package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

	// Create the mirroring
	if repository.Spec.Mirroring != nil {
		logf.Log.Info("[Repository Service] Creating mirroring repository", "name", repository.Spec.Name)

		// Change repoState for enable mirroring
		err = s.QuayClientMgr.Get().ChangeRepositoryState(repository.Spec.Name, "MIRROR")
		if err != nil {
			logf.Log.Error(err, "[Repository Service] Error when updating repository state", "name", repository.Spec.Name, "state", "MIRROR")
			return err
		} else {
			logf.Log.Info("[Repository Service] Repository successfully updated state", "name", repository.Spec.Name, "state", "MIRROR")
		}

		externalRegistryUsername, externalRegistryPassword, err := s.resolveMirroringCredentials(ctx, repository)
		if err != nil {
			return err
		}

		// Prepare the model
		repositoryMirrorToCreate := quay.CreateRepositoryMirror{
			ExternalReference:        repository.Spec.Mirroring.RegistryLocation,
			ExternalRegistryUsername: externalRegistryUsername,
			ExternalRegistryPassword: externalRegistryPassword,
			SyncInterval:             syncIntervalInSeconds(repository.Spec.Mirroring.SyncInterval),
			SyncStartDate:            setDefaultDateTimeNow(&repository.Spec.Mirroring.StartDate),
			RobotUsername:            repository.Spec.Mirroring.RobotUser,
			ExternalRegistryConfig: quay.ExternalRegistryConfig{
				VerifyTLS:      repository.Spec.Mirroring.VerifyTLS,
				UnsignedImages: repository.Spec.Mirroring.AcceptUnsignedImages,
				Proxy: quay.Proxy{
					HttpProxy:  repository.Spec.Mirroring.HttpProxy,
					HttpsProxy: repository.Spec.Mirroring.HttpsProxy,
					NoProxy:    repository.Spec.Mirroring.NoProxy,
				},
			},
			RootRule: quay.RootRule{
				RuleKind:  "tag_glob_csv", // I think tag_glob_csv is by default
				RuleValue: repository.Spec.Mirroring.Tags,
			},
		}

		mirrorCreated, err := s.QuayClientMgr.Get().CreateRepositoryMirror(repository.Spec.Name, repositoryMirrorToCreate)
		if err != nil {
			logf.Log.Error(err, "[Repository Service] Error when creating repository mirror", "name", repository.Spec.Name)
			return err
		} else {
			logf.Log.Info("[Repository Service] Created repository mirror", "name", repository.Spec.Name, "mirror", mirrorCreated)
		}
	}

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

func syncIntervalInSeconds(syncInterval string) int {
	value := strings.TrimSpace(syncInterval)
	if value == "" {
		return 0
	}

	lastChar := value[len(value)-1]
	if lastChar >= '0' && lastChar <= '9' {
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds < 0 {
			return 0
		}
		return seconds
	}

	quantity, err := strconv.Atoi(value[:len(value)-1])
	if err != nil || quantity < 0 {
		return 0
	}

	switch strings.ToLower(string(lastChar)) {
	case "s":
		return quantity
	case "m":
		return quantity * 60
	case "h":
		return quantity * 60 * 60
	case "d":
		return quantity * 60 * 60 * 24
	case "w":
		return quantity * 60 * 60 * 24 * 7
	default:
		return 0
	}
}

// resolveMirroringCredentials returns optional username/password for Quay mirroring.
// Values can be provided inline in the CR or sourced from Secret/ConfigMap refs.
func (s *RepositoryService) resolveMirroringCredentials(ctx context.Context, repository *quayiov1alpha.Repository) (*string, *string, error) {
	credentials := repository.Spec.Mirroring.Credentials

	username := strings.TrimSpace(credentials.Username)
	if username == "" && credentials.UsernameFrom != nil {
		resolvedUsername, err := s.resolveCredentialSourceValue(ctx, repository.Namespace, credentials.UsernameFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve mirroring username for repository %q: %w", repository.Name, err)
		}
		username = strings.TrimSpace(resolvedUsername)
	}

	password := credentials.Password
	if password == "" && credentials.PasswordFrom != nil {
		resolvedPassword, err := s.resolveCredentialSourceValue(ctx, repository.Namespace, credentials.PasswordFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve mirroring password for repository %q: %w", repository.Name, err)
		}
		password = resolvedPassword
	}

	var usernamePtr *string
	if username != "" {
		usernamePtr = &username
	}

	var passwordPtr *string
	if password != "" {
		passwordPtr = &password
	}

	return usernamePtr, passwordPtr, nil
}

func (s *RepositoryService) resolveCredentialSourceValue(ctx context.Context, namespace string, source *quayiov1alpha.RepositoryCredentialSource) (string, error) {
	if source == nil {
		return "", nil
	}

	if source.SecretKeyRef != nil {
		var secret corev1.Secret
		err := s.KubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: source.SecretKeyRef.Name}, &secret)
		if err != nil {
			return "", fmt.Errorf("unable to read secret %q: %w", source.SecretKeyRef.Name, err)
		}

		value, ok := secret.Data[source.SecretKeyRef.Key]
		if !ok {
			return "", fmt.Errorf("key %q not found in secret %q", source.SecretKeyRef.Key, source.SecretKeyRef.Name)
		}
		return string(value), nil
	}

	if source.ConfigMapKeyRef != nil {
		var configMap corev1.ConfigMap
		err := s.KubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: source.ConfigMapKeyRef.Name}, &configMap)
		if err != nil {
			return "", fmt.Errorf("unable to read configmap %q: %w", source.ConfigMapKeyRef.Name, err)
		}

		value, ok := configMap.Data[source.ConfigMapKeyRef.Key]
		if !ok {
			return "", fmt.Errorf("key %q not found in configmap %q", source.ConfigMapKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
		return value, nil
	}

	return "", nil
}
