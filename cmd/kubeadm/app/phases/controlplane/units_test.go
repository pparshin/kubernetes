/*
Copyright 2021 The Kubernetes Authors.

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

package controlplane

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	testutil "k8s.io/kubernetes/cmd/kubeadm/test"
)

func TestGetServiceUnitSpecs(t *testing.T) {
	cfg := &kubeadmapi.ClusterConfiguration{
		KubernetesVersion: "v1.9.0",
	}

	specs := getServiceUnitSpecs(cfg, &kubeadmapi.APIEndpoint{})

	var tests = []struct {
		name          string
		cmd           string
		componentName string
	}{
		{
			name:          "KubeAPIServer",
			cmd:           "kube-apiserver",
			componentName: kubeadmconstants.KubeAPIServer,
		},
		{
			name:          "KubeControllerManager",
			cmd:           "kube-controller-manager",
			componentName: kubeadmconstants.KubeControllerManager,
		},
		{
			name:          "KubeScheduler",
			cmd:           "kube-scheduler",
			componentName: kubeadmconstants.KubeScheduler,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if assert.Contains(t, specs, tc.componentName) {
				spec := specs[tc.componentName]
				assert.Equal(t, tc.cmd, spec.Service.ExecStartCmd)
			}
		})
	}
}

func TestCreateServiceUnitFiles(t *testing.T) {
	var tests = []struct {
		name       string
		components []string
	}{
		{
			name:       "KubeAPIServer KubeAPIServer KubeScheduler",
			components: []string{kubeadmconstants.KubeAPIServer, kubeadmconstants.KubeControllerManager, kubeadmconstants.KubeScheduler},
		},
		{
			name:       "KubeAPIServer",
			components: []string{kubeadmconstants.KubeAPIServer},
		},
		{
			name:       "KubeControllerManager",
			components: []string{kubeadmconstants.KubeControllerManager},
		},
		{
			name:       "KubeScheduler",
			components: []string{kubeadmconstants.KubeScheduler},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpdir := testutil.SetupTempDir(t)
			defer os.RemoveAll(tmpdir)

			cfg := &kubeadmapi.ClusterConfiguration{
				KubernetesVersion: "v1.9.0",
			}

			unitsDir := filepath.Join(tmpdir, "units")
			err := CreateServiceUnitFiles(unitsDir, cfg, &kubeadmapi.APIEndpoint{}, tt.components...)
			require.NoError(t, err)

			testutil.AssertFilesCount(t, unitsDir, len(tt.components))

			for _, fileName := range tt.components {
				assert.FileExists(t, kubeadmconstants.GetSystemUnitFilepath(fileName, unitsDir))
			}
		})
	}
}
