package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RobotAccountUIDLabel            = "quay.io/robot-account-uid"
	RobotAccountNameAnnotation      = "quay.io/robot-account-name"
	RobotAccountNamespaceAnnotation = "quay.io/robot-account-namespace"
)

type RobotAccountService struct {
	KubeClient    client.Client
	QuayClientMgr *quay.ClientManager
	Scheme        *runtime.Scheme
}

func NewRobotAccountService(kubeClient client.Client, quayClientMgr *quay.ClientManager, scheme *runtime.Scheme) *RobotAccountService {
	return &RobotAccountService{
		KubeClient:    kubeClient,
		QuayClientMgr: quayClientMgr,
		Scheme:        scheme,
	}
}

func (s *RobotAccountService) Reconcile(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount) (bool, error) {
	// Ensure Quay client is initialized
	if s.QuayClientMgr.Get() == nil {
		return false, errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	// Search if the robot account exists on Quay
	var existingRobotAccount quay.RobotAccount
	var existingErr error
	if strings.ToLower(robotAccount.Spec.Owner.Kind) == "user" {
		// Quay api do not allowed to get a robot account from a specific user (by his name), so all user robot accounts, will be created above the api key user/robotAccount
		existingRobotAccount, existingErr = s.QuayClientMgr.Get().GetUserRobotAccount(robotAccount.Spec.Name)
	} else if strings.ToLower(robotAccount.Spec.Owner.Kind) == "organization" {
		existingRobotAccount, existingErr = s.QuayClientMgr.Get().GetOrganizationRobotAccount(robotAccount.Spec.Owner.Name, robotAccount.Spec.Name)
	} else {
		return false, fmt.Errorf("unsupported robot account owner kind %q", robotAccount.Spec.Owner.Kind)
	}

	if existingErr != nil {
		if errors.Is(existingErr, quay.ErrRobotAccountNotFound) {
			// Robot account not found, create it
			existingRobotAccount, existingErr = s.create(robotAccount)
			if existingErr != nil {
				return false, existingErr
			}
		} else {
			// other errors, raise them
			return false, existingErr
		}
	}

	if err := s.reconcileSecret(ctx, robotAccount, &existingRobotAccount); err != nil {
		return false, err
	}
	if existingRobotAccount.Description != robotAccount.Spec.Description {
		return false, errors.New("updating an existing Quay robot account description is not supported")
	}
	return false, nil
}

func (s *RobotAccountService) Delete(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount) error {
	logf.Log.Info("[Robot Account Service] Deleting robot account", "name", robotAccount.Spec.Name)

	if s.QuayClientMgr.Get() == nil {
		return errors.New("Quay client not initialized; ensure QuayConfig is configured")
	}

	var deleteErr error

	if strings.ToLower(robotAccount.Spec.Owner.Kind) == "user" {
		deleteErr = s.QuayClientMgr.Get().DeleteUserRobotAccount(robotAccount.Spec.Name)
	} else if strings.ToLower(robotAccount.Spec.Owner.Kind) == "organization" {
		deleteErr = s.QuayClientMgr.Get().DeleteOrganizationRobotAccount(robotAccount.Spec.Owner.Name, robotAccount.Spec.Name)
	} else {
		return fmt.Errorf("unsupported robot account owner kind %q", robotAccount.Spec.Owner.Kind)
	}

	if deleteErr != nil && !errors.Is(deleteErr, quay.ErrNotFound) {
		logf.Log.Error(deleteErr, "[Robot Account Service] Error when deleting robot account", "name", robotAccount.Spec.Name)
		return deleteErr
	}

	return s.deleteObsoleteSecrets(ctx, robotAccount, nil)
}

func (s *RobotAccountService) create(robotAccount *quayiov1alpha.RobotAccount) (quay.RobotAccount, error) {
	logf.Log.Info("[Robot Account Service] Creating robot account", "name", robotAccount.Spec.Name, "kind", robotAccount.Spec.Owner.Kind)

	// Prepare model
	robotAccountToCreate := quay.CreateRobotAccount{
		Description: robotAccount.Spec.Description,
	}

	// Create on Quay
	var createErr error
	var createdRobotAccount quay.RobotAccount
	if strings.ToLower(robotAccount.Spec.Owner.Kind) == "user" {
		createdRobotAccount, createErr = s.QuayClientMgr.Get().CreateUserRobotAccount(robotAccount.Spec.Name, robotAccountToCreate)
	} else if strings.ToLower(robotAccount.Spec.Owner.Kind) == "organization" {
		createdRobotAccount, createErr = s.QuayClientMgr.Get().CreateOrganizationRobotAccount(robotAccount.Spec.Owner.Name, robotAccount.Spec.Name, robotAccountToCreate)
	} else {
		return quay.RobotAccount{}, fmt.Errorf("unsupported robot account owner kind %q", robotAccount.Spec.Owner.Kind)
	}

	if createErr != nil {
		logf.Log.Error(createErr, "[Robot Account Service] Error when creating robot account", "name", robotAccount.Spec.Name, "kind", robotAccount.Spec.Owner.Kind)
		return quay.RobotAccount{}, createErr
	} else {
		logf.Log.Info("[Robot Account Service] Created robot account", "name", robotAccount.Spec.Name, "kind", robotAccount.Spec.Owner.Kind)
	}

	return createdRobotAccount, nil
}

func (s *RobotAccountService) reconcileSecret(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount, quayRobot *quay.RobotAccount) error {
	if robotAccount.Spec.Secret == nil {
		return s.deleteObsoleteSecrets(ctx, robotAccount, nil)
	}
	if robotAccount.UID == "" {
		return errors.New("robot account UID must not be empty")
	}
	if quayRobot.Name == "" || quayRobot.Token == "" {
		return errors.New("Quay robot account response must include name and token")
	}

	// Select the namespace
	namespace := robotAccount.Spec.Secret.Namespace
	if namespace == "" {
		namespace = robotAccount.Namespace // by default
	}
	secretName := robotAccount.Spec.Secret.Name
	if secretName == "" {
		secretName = strings.NewReplacer("+", "-", "_", "-").Replace(strings.ToLower(quayRobot.Name)) + "-pull-secret"
	}
	if problems := validation.IsDNS1123Subdomain(secretName); len(problems) > 0 {
		return fmt.Errorf("invalid secret name %q: %s", secretName, strings.Join(problems, "; "))
	}
	key := types.NamespacedName{Name: secretName, Namespace: namespace}

	dockerConfig, err := createDockerCfg(quayRobot.Name, quayRobot.Token, s.QuayClientMgr.Get().BaseURL)
	if err != nil {
		return err
	}

	config, err := json.Marshal(dockerConfig)
	if err != nil {
		return fmt.Errorf("marshal Docker config: %w", err)
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace}}
	_, err = controllerutil.CreateOrUpdate(ctx, s.KubeClient, secret, func() error {
		if secret.ResourceVersion != "" && (secret.Labels[RobotAccountUIDLabel] != string(robotAccount.UID) ||
			secret.Annotations[RobotAccountNameAnnotation] != robotAccount.Name ||
			secret.Annotations[RobotAccountNamespaceAnnotation] != robotAccount.Namespace) {
			return fmt.Errorf("secret %s is not managed by this robot account", key)
		}
		if namespace == robotAccount.Namespace {
			if err := controllerutil.SetControllerReference(robotAccount, secret, s.Scheme); err != nil {
				return fmt.Errorf("set secret owner: %w", err)
			}
		}
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		secret.Labels[RobotAccountUIDLabel] = string(robotAccount.UID)
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}
		secret.Annotations[RobotAccountNameAnnotation] = robotAccount.Name
		secret.Annotations[RobotAccountNamespaceAnnotation] = robotAccount.Namespace
		secret.Type = corev1.SecretTypeDockerConfigJson
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[corev1.DockerConfigJsonKey] = config
		return nil
	})
	if err != nil {
		return fmt.Errorf("reconcile secret %s: %w", key, err)
	}
	return s.deleteObsoleteSecrets(ctx, robotAccount, &key)
}

func (s *RobotAccountService) deleteObsoleteSecrets(ctx context.Context, robotAccount *quayiov1alpha.RobotAccount, keep *types.NamespacedName) error {
	if robotAccount.UID == "" {
		return errors.New("robot account UID must not be empty")
	}
	var secrets corev1.SecretList
	if err := s.KubeClient.List(ctx, &secrets, client.MatchingLabels{RobotAccountUIDLabel: string(robotAccount.UID)}); err != nil {
		return fmt.Errorf("list robot account secrets: %w", err)
	}
	for i := range secrets.Items {
		secret := &secrets.Items[i]
		if secret.Annotations[RobotAccountNameAnnotation] != robotAccount.Name ||
			secret.Annotations[RobotAccountNamespaceAnnotation] != robotAccount.Namespace {
			continue
		}
		if keep != nil && secret.Name == keep.Name && secret.Namespace == keep.Namespace {
			continue
		}
		if err := s.KubeClient.Delete(ctx, secret, client.Preconditions{UID: &secret.UID, ResourceVersion: &secret.ResourceVersion}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete obsolete secret %s/%s: %w", secret.Namespace, secret.Name, err)
		}
	}
	return nil
}

/*
createDockerCfg Create the dockercfg (dockerconfig.json)
*/
func createDockerCfg(robotAccountName string, token string, quayURL string) (DockerConfig, error) {
	parsedURL, err := url.Parse(quayURL)
	if err != nil {
		return DockerConfig{}, fmt.Errorf("parse Quay URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return DockerConfig{}, fmt.Errorf("Quay URL must include a scheme and host")
	}

	quayDomain := parsedURL.Host
	auth := base64.StdEncoding.EncodeToString([]byte(robotAccountName + ":" + token))
	dockerConfig := DockerConfig{
		Auths: map[string]DockerAuthConfig{
			quayDomain: {Auth: auth},
		},
	}

	return dockerConfig, nil
}
