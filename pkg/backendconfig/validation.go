/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package backendconfig

import (
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	backendconfigv1beta1 "k8s.io/ingress-gce/pkg/apis/backendconfig/v1beta1"
)

const (
	OAuthClientIDKey     = "client_id"
	OAuthClientSecretKey = "client_secret"
)

func Validate(kubeClient kubernetes.Interface, beConfig *backendconfigv1beta1.BackendConfig) error {
	if beConfig == nil {
		return nil
	}
	return validateIAP(kubeClient, beConfig)
}

// TODO(rramkumar): Return errors as constants so that the unit tests can distinguish
// between which error is returned.
func validateIAP(kubeClient kubernetes.Interface, beConfig *backendconfigv1beta1.BackendConfig) error {
	// If IAP settings are not found or IAP is not enabled then don't bother continuing.
	if beConfig.Spec.Iap == nil || beConfig.Spec.Iap.Enabled == false {
		return nil
	}
	// If necessary, get the OAuth credentials stored in the K8s secret.
	if beConfig.Spec.Iap.OAuthClientCredentials != nil && beConfig.Spec.Iap.OAuthClientCredentials.SecretName != "" {
		secretName := beConfig.Spec.Iap.OAuthClientCredentials.SecretName
		secret, err := kubeClient.Core().Secrets(beConfig.Namespace).Get(secretName, meta_v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error retrieving secret %v: %v", secretName, err)
		}
		clientID, ok := secret.Data[OAuthClientIDKey]
		if !ok {
			return fmt.Errorf("secret %v missing %v data", secretName, OAuthClientIDKey)
		}
		clientSecret, ok := secret.Data[OAuthClientSecretKey]
		if !ok {
			return fmt.Errorf("secret %v missing %v data'", secretName, OAuthClientSecretKey)
		}
		beConfig.Spec.Iap.OAuthClientCredentials.ClientID = string(clientID)
		beConfig.Spec.Iap.OAuthClientCredentials.ClientSecret = string(clientSecret)
	}

	if beConfig.Spec.Cdn != nil && beConfig.Spec.Cdn.Enabled {
		return fmt.Errorf("iap and cdn cannot be enabled at the same time")
	}
	return nil
}
