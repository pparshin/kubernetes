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

package initsystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

func TestWriteUnitToDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	component := kubeadmconstants.KubeScheduler
	unit := UnitSpec{
		Description:   "Kubernetes Scheduler",
		Documentation: "https://github.com/kubernetes/kubernetes",
		Service: UnitService{
			ExecStartCmd: []string{
				"kube-scheduler",
				"--bind-address=127.0.0.1",
				"--leader-elect=true",
				"--kubeconfig=scheduler.conf",
				"--authentication-kubeconfig=scheduler.conf",
				"--authorization-kubeconfig=scheduler.conf",
			},
			Restart:    UnitRestartOnFailure,
			RestartSec: DefaultUnitRestartSec,
		},
		Install: UnitInstall{
			WantedBy: MultiUserTarget,
		},
	}

	unitsDir := filepath.Join(tmpdir, "units")
	err = WriteUnitToDisk(component, unitsDir, unit)
	require.NoError(t, err)

	path := kubeadmconstants.GetSystemUnitFilepath(component, unitsDir)
	want := `
[Unit]
Description=Kubernetes Scheduler
Documentation=https://github.com/kubernetes/kubernetes

[Service]
ExecStart=kube-scheduler --bind-address=127.0.0.1 --leader-elect=true --kubeconfig=scheduler.conf --authentication-kubeconfig=scheduler.conf --authorization-kubeconfig=scheduler.conf 
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

	got, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.Equal(t, want, string(got))
}
