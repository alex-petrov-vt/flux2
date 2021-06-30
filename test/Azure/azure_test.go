package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
)

func TestAzureE2E(t *testing.T) {
	ctx := context.TODO()

	// Apply terraform
	terraformOpts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./terraform",
	})
	//defer terraform.Destroy(t, terraformOptions)
	//terraform.InitAndApply(t, terraformOpts)
	kubeConfig := terraform.Output(t, terraformOpts, "aks_kube_config")
	aksHost := terraform.Output(t, terraformOpts, "aks_host")
	aksCert := terraform.Output(t, terraformOpts, "aks_client_certificate")
	aksKey := terraform.Output(t, terraformOpts, "aks_client_key")
	aksCa := terraform.Output(t, terraformOpts, "aks_cluster_ca_certificate")

	// Generate kubeconfig for cluster
	tmpDir, err := ioutil.TempDir("", "*-azure-e2e")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	kubeConfigPath := fmt.Sprintf("%s/kubeconfig", tmpDir)
	os.WriteFile(kubeConfigPath, []byte(kubeConfig), 0750)

	// Install Flux using CLI
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, "flux", "install", "--components-extra", "image-reflector-controller,image-automation-controller", "--kubeconfig", kubeConfigPath)
	_, err = cmd.Output()
	require.NoError(t, err)

	// Create kubernetes client
	err = sourcev1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	err = kustomizev1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	kubeCfg := &rest.Config{
		Host: aksHost,
		TLSClientConfig: rest.TLSClientConfig{
			CertData: []byte(aksCert),
			KeyData:  []byte(aksKey),
			CAData:   []byte(aksCa),
		},
	}
	kubeClient, err := client.New(kubeCfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)

	// Bootstrap cluster
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		StringData: map[string]string{},
	}
	err = kubeClient.Create(ctx, &secret, &client.CreateOptions{})
	require.NoError(t, err)

	source := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		Spec: sourcev1.GitRepositorySpec{
			GitImplementation: sourcev1.LibGit2Implementation,
			Reference: &sourcev1.GitRepositoryRef{
				Branch: "main",
			},
			SecretRef: &meta.LocalObjectReference{
				Name: "flux-system",
			},
			URL: "ssh://ssh.dev.azure.com/v3/flux-e2e/flux/flux",
		},
	}
	err = kubeClient.Create(ctx, source, &client.CreateOptions{})
	require.NoError(t, err)

	kustomization := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "./cluster/prod",
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind:      sourcev1.GitRepositoryKind,
				Name:      "flux-system",
				Namespace: "flux-system",
			},
		},
	}
	err = kubeClient.Create(ctx, kustomization, &client.CreateOptions{})
	require.NoError(t, err)
}
