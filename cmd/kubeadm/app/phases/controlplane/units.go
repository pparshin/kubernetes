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
	"github.com/pkg/errors"

	"k8s.io/klog/v2"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/initsystem"
)

func getServiceUnitSpecs(cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint) map[string]initsystem.UnitSpec {
	unitSpecs := map[string]initsystem.UnitSpec{
		kubeadmconstants.KubeAPIServer: {
			Description:   "Kubernetes API Server",
			Documentation: "https://github.com/kubernetes/kubernetes",
			Service: initsystem.UnitService{
				ExecStartCmd:  getAPIServerCommand(cfg, endpoint),
				Restart:       initsystem.UnitRestartAlways,
				RestartSec:    initsystem.DefaultUnitRestartSec,
			},
			Install: initsystem.UnitInstall{
				WantedBy: initsystem.MultiUserTarget,
			},
		},
		kubeadmconstants.KubeControllerManager: {
			Description:   "Kubernetes Controller Manager",
			Documentation: "https://github.com/kubernetes/kubernetes",
			Service: initsystem.UnitService{
				ExecStartCmd:  getControllerManagerCommand(cfg),
				Restart:       initsystem.UnitRestartAlways,
				RestartSec:    initsystem.DefaultUnitRestartSec,
			},
			Install: initsystem.UnitInstall{
				WantedBy: initsystem.MultiUserTarget,
			},
		},
		kubeadmconstants.KubeScheduler: {
			Description:   "Kubernetes Scheduler",
			Documentation: "https://github.com/kubernetes/kubernetes",
			Service: initsystem.UnitService{
				ExecStartCmd:  getSchedulerCommand(cfg),
				Restart:       initsystem.UnitRestartAlways,
				RestartSec:    initsystem.DefaultUnitRestartSec,
			},
			Install: initsystem.UnitInstall{
				WantedBy: initsystem.MultiUserTarget,
			},
		},
	}

	return unitSpecs
}

// CreateServiceUnitFiles creates all the requested service unit files.
func CreateServiceUnitFiles(unitsDir string, cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint, componentNames ...string) error {
	specs := getServiceUnitSpecs(cfg, endpoint)

	for _, componentName := range componentNames {
		spec, exists := specs[componentName]
		if !exists {
			return errors.Errorf("couldn't retrieve service unit for %q", componentName)
		}

		if err := initsystem.WriteUnitToDisk(componentName, unitsDir, spec); err != nil {
			return errors.Wrapf(err, "failed to create service unit file for %q", componentName)
		}

		klog.V(1).Infof("[control-plane] wrote service unit for component %q to %q\n", componentName, kubeadmconstants.GetSystemUnitFilepath(componentName, unitsDir))
	}

	return nil
}

// RunServices runs Kubernetes components as system daemons.
func RunServices(componentNames ...string) error {
	sys, err := initsystem.GetInitSystem()
	if err != nil {
		return err
	}

	for _, component := range componentNames {
		err = sys.ServiceStart(component)
		if err != nil {
			return err
		}
	}

	return nil
}
