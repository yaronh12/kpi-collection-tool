package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	THANOS_ROUTE_API_PATH = "/apis/route.openshift.io/v1/namespaces/openshift-monitoring/routes/thanos-querier"
)

// setupKubeconfigAuth sets up authentication using kubeconfig file
// and discovers Thanos URL and creates service account token
func SetupKubeconfigAuth(kubeconfig string) (string, string, error) {
	clientset, err := setupKubernetesClient(kubeconfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to setup Kubernetes client: %v", err)
	}

	thanosURL, err := getThanosURL(clientset)
	if err != nil {
		return "", "", fmt.Errorf("failed to get Thanos URL: %v", err)
	}

	bearerToken, err := createServiceAccountToken(clientset)
	if err != nil {
		return "", "", fmt.Errorf("failed to create service account token: %v", err)
	}

	return thanosURL, bearerToken, nil
}

// setupKubernetesClient creates a Kubernetes clientset from kubeconfig
func setupKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return clientset, nil
}

// getThanosURL retrieves the Thanos querier route hostname from OpenShift
// Equivalent to: oc get route thanos-querier -n openshift-monitoring -o jsonpath='{.spec.host}'
func getThanosURL(clientset *kubernetes.Clientset) (string, error) {
	routes, err := clientset.RESTClient().
		Get().
		AbsPath(THANOS_ROUTE_API_PATH).
		DoRaw(context.TODO())

	if err != nil {
		return "", fmt.Errorf("failed to get thanos-querier route: %v", err)
	}

	var route route
	if err := json.Unmarshal(routes, &route); err != nil {
		return "", fmt.Errorf("failed to parse route response: %v", err)
	}

	if route.Spec.Host != "" {
		return route.Spec.Host, nil
	}

	return "", fmt.Errorf("failed to extract host from route spec")
}

// createServiceAccountToken creates a service account token for authentication
// Equivalent to: oc create token telemeter-client -n openshift-monitoring --duration=10h
func createServiceAccountToken(clientset *kubernetes.Clientset) (string, error) {
	tokenRequest := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			ExpirationSeconds: int64Ptr(36000), // 10 hours = 36000 seconds
		},
	}

	result, err := clientset.CoreV1().ServiceAccounts("openshift-monitoring").
		CreateToken(context.TODO(), "telemeter-client", tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create service account token: %v", err)
	}

	return result.Status.Token, nil
}

// int64Ptr returns a pointer to an int64 value
func int64Ptr(i int64) *int64 {
	return &i
}
