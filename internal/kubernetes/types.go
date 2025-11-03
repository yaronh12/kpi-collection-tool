package kubernetes

import (
	"context"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Route represents the OpenShift route object structure
type route struct {
	Spec struct {
		Host string `json:"host"`
	} `json:"spec"`
}

// K8sClient is an interface for Kubernetes operations we need
type K8sClient interface {
	GetRouteRaw(ctx context.Context, path string) ([]byte, error)
	CreateServiceAccountToken(ctx context.Context, namespace, serviceAccount string, tokenRequest *authv1.TokenRequest) (*authv1.TokenRequest, error)
}

// k8sClientImpl implements K8sClient using a real Kubernetes clientset
type k8sClientImpl struct {
	clientset *kubernetes.Clientset
}

func (k *k8sClientImpl) GetRouteRaw(ctx context.Context, path string) ([]byte, error) {
	return k.clientset.RESTClient().
		Get().
		AbsPath(path).
		DoRaw(ctx)
}

func (k *k8sClientImpl) CreateServiceAccountToken(ctx context.Context, namespace, serviceAccount string, tokenRequest *authv1.TokenRequest) (*authv1.TokenRequest, error) {
	return k.clientset.CoreV1().ServiceAccounts(namespace).
		CreateToken(ctx, serviceAccount, tokenRequest, metav1.CreateOptions{})
}
