package e2e

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/operator-framework/catalogd/api/core/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestE2E(t *testing.T) {
	catalog := &v1alpha1.ClusterCatalog{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-catalog",
		},
		Spec: v1alpha1.ClusterCatalogSpec{
			Source: v1alpha1.CatalogSource{
				Type: v1alpha1.SourceTypeImage,
				Image: &v1alpha1.ImageSource{
					Ref:                   os.Getenv("CATALOG_IMG"),
					InsecureSkipTLSVerify: true,
				},
			},
		},
	}

	cfg := ctrl.GetConfigOrDie()
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	cli, err := client.New(cfg, client.Options{Scheme: scheme})
	require.NoError(t, err)

	err = cli.Create(context.Background(), catalog)
	require.NoError(t, err)

	var tests = []struct {
		name           string
		command        *exec.Cmd
		expectedOutput string
		expectedError  bool
	}{
		{
			name:    "list all content",
			command: exec.Command("../../kubectl-catalogd", "list"),
			expectedOutput: ` test-catalog  olm.package  prometheus
 test-catalog  olm.channel prometheus alpha
 test-catalog  olm.channel prometheus beta
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.1
 test-catalog  olm.bundle prometheus prometheus-operator.1.2.0
 test-catalog  olm.bundle prometheus prometheus-operator.2.0.0
 test-catalog  olm.package  plain
 test-catalog  olm.channel plain beta
 test-catalog  olm.bundle plain plain.0.1.0
`,
		},
		{
			name:    "list all content with schema olm.package",
			command: exec.Command("../../kubectl-catalogd", "list", "--schema", "olm.package"),
			expectedOutput: ` test-catalog  olm.package  prometheus
 test-catalog  olm.package  plain
`,
		},
		{
			name:    "list all content with schema olm.bundle and package prometheus",
			command: exec.Command("../../kubectl-catalogd", "list", "--schema", "olm.bundle", "--package", "prometheus"),
			expectedOutput: ` test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.1
 test-catalog  olm.bundle prometheus prometheus-operator.1.2.0
 test-catalog  olm.bundle prometheus prometheus-operator.2.0.0
`,
		},
		{
			name:    "list all content with schema olm.bundle, package prometheus, and name prometheus-operator.1.0.0",
			command: exec.Command("../../kubectl-catalogd", "list", "--schema", "olm.bundle", "--package", "prometheus", "--name", "prometheus-operator.1.0.0"),
			expectedOutput: ` test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
`,
		},
		{
			name:    "search for content with name containing 'prom'",
			command: exec.Command("../../kubectl-catalogd", "search", "prom"),
			expectedOutput: ` test-catalog  olm.package  prometheus
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.1
 test-catalog  olm.bundle prometheus prometheus-operator.1.2.0
 test-catalog  olm.bundle prometheus prometheus-operator.2.0.0
`,
		},
		{
			name:    "search for content with name containing 'prom' and schema olm.package",
			command: exec.Command("../../kubectl-catalogd", "search", "prom", "--schema", "olm.package"),
			expectedOutput: ` test-catalog  olm.package  prometheus
`,
		},
		{
			name:    "search for content with name containing 'p' and schema olm.bundle and package plain",
			command: exec.Command("../../kubectl-catalogd", "search", "p", "--schema", "olm.bundle", "--package", "plain"),
			expectedOutput: ` test-catalog  olm.bundle plain plain.0.1.0
`,
		},
		{
			name:    "inspect olm.package with name prometheus",
			command: exec.Command("../../kubectl-catalogd", "inspect", "olm.package", "prometheus"),
			expectedOutput: `{
  "defaultChannel": "beta",
  "name": "prometheus",
  "schema": "olm.package"
}`,
		},
		{
			name:    "inspect olm.package with name prometheus and output yaml",
			command: exec.Command("../../kubectl-catalogd", "inspect", "olm.package", "prometheus", "--output", "yaml"),
			expectedOutput: `defaultChannel: beta
name: prometheus
schema: olm.package
`,
		},
		{
			name:    "inspect olm.channel with name beta and package plain",
			command: exec.Command("../../kubectl-catalogd", "inspect", "olm.channel", "beta", "--package", "plain", "--output", "yaml"),
			expectedOutput: `entries:
- name: plain.0.1.0
name: beta
package: plain
schema: olm.channel
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.command.CombinedOutput()
			if tt.expectedError {
				require.Error(t, err)
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedOutput, string(output))
		})
	}
}
