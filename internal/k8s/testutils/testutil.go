// +build test

package testutils

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
)

var s = k8sRuntime.NewScheme()

func TestK8SClient() (client.Client, func(), error) {
	if err := scheme.AddToScheme(s); err != nil {
		return nil, nil, fmt.Errorf("failed to create new scheme: %w", err)
	}

	if err := apiv1.AddToScheme(s); err != nil {
		return nil, nil, fmt.Errorf("faileld to add bootes scheme: %w", err)
	}

	_, file, _, _ := runtime.Caller(0)
	testEnv := envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join(path.Dir(file), "..", "..", "..", "kubernetes", "kpt", "crd")},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("faileld to start test env: %w", err)
	}

	cli, err := client.New(cfg, client.Options{
		Scheme: s,
	})
	if err != nil {
		err = fmt.Errorf("faileld to create controller-runtime client: %w", err)

		if nerr := testEnv.Stop(); err != nil {
			err = fmt.Errorf("failed to stop test env: %w", nerr)
		}

		return nil, nil, err
	}

	return cli, func() {
		if err := testEnv.Stop(); err != nil {
			panic(fmt.Sprintf("failed to stop envtest instance: %s", err))
		}
	}, nil
}

func NewNamespace(t *testing.T, ctx context.Context, cli client.Client) string {
	t.Helper()

	name := uuid.New().String()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := cli.Create(ctx, ns); err != nil {
		t.Fatalf("failed to create namespace: %s", err)
	}

	return name
}
