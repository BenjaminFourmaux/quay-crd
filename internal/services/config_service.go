package services

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	quayiov1alpha "quay-crd/api/v1alpha"
	"quay-crd/internal/services/quay"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigService struct {
	KubeClient client.Client
	QuayClient *quay.ClientManager
}

/*
NewConfigService is the constructor for ConfigService
*/
func NewConfigService(kubeClient client.Client) *ConfigService {
	return &ConfigService{
		KubeClient: kubeClient,
		QuayClient: quay.NewClientManager(nil),
	}
}

func (s *ConfigService) Reconcile(ctx context.Context, quayConfig *quayiov1alpha.QuayConfig) error {
	// 1. Get secret data from SecretRef
	var secret corev1.Secret

	err := s.KubeClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: quayConfig.Namespace,
			Name:      quayConfig.Spec.Credentials.Name,
		},
		&secret,
	)

	if err != nil {
		return fmt.Errorf("failed to get Quay credentials secret: %w", err)
	}

	token, ok := secret.Data[quayConfig.Spec.Credentials.Key]
	if !ok {
		return fmt.Errorf(
			"key %q not found in secret %q",
			quayConfig.Spec.Credentials.Key,
			quayConfig.Spec.Credentials.Name,
		)
	}

	// 2. Initialize Quay Client with specified creds
	quayClient := quay.NewClient(quayConfig.Spec.URL, string(token))

	// 3. Ping Quay to validate connectivity
	if err = quayClient.Ping(); err != nil {
		return s.setNotReady(
			ctx,
			quayConfig,
			"ConnectionFailed",
			fmt.Sprintf("Unable to connect to Quay: %v", err),
		)
	} else {
		_ = s.setReady(
			ctx,
			quayConfig,
			"ConnectionSucceeded",
			"Successfully connected to Quay",
		)
	}

	s.QuayClient.Set(quayClient)

	return nil
}

func (s *ConfigService) setNotReady(
	ctx context.Context,
	quayConfig *quayiov1alpha.QuayConfig,
	reason string,
	message string,
) error {
	quayConfig.Status.Ready = false

	meta.SetStatusCondition(&quayConfig.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
	return s.KubeClient.Status().Update(ctx, quayConfig)
}

func (s *ConfigService) setReady(
	ctx context.Context,
	quayConfig *quayiov1alpha.QuayConfig,
	reason string,
	message string,
) error {
	quayConfig.Status.Ready = true

	meta.SetStatusCondition(&quayConfig.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
	return s.KubeClient.Status().Update(ctx, quayConfig)
}
