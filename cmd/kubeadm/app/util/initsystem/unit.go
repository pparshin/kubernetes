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
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/pkg/errors"

	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

type UnitRestartMode string

const (
	UnitRestartNo        UnitRestartMode = "no"
	UnitRestartOnFailure UnitRestartMode = "on-failure"
	UnitRestartOnSuccess UnitRestartMode = "on-success"
	UnitRestartOnAbort   UnitRestartMode = "on-abort"
	UnitRestartAlways    UnitRestartMode = "always"
)

const (
	MultiUserTarget = "multi-user.target"
)

const (
	DefaultUnitRestartSec = 5
)

type UnitService struct {
	ExecStartCmd []string
	Restart      UnitRestartMode
	RestartSec   uint
}

type UnitInstall struct {
	Alias      string
	WantedBy   string
	RequiredBy string
}

type UnitSpec struct {
	Description   string
	Documentation string
	Service       UnitService
	Install       UnitInstall
}

const (
	unitTemplate = `
[Unit]
Description={{.Description}}
Documentation={{.Documentation}}

[Service]
ExecStart={{range .Service.ExecStartCmd}}{{.}} {{end}}
Restart={{.Service.Restart}}
RestartSec={{.Service.RestartSec}}

[Install]
{{- if .Install.Alias}}
Alias={{.Install.Alias}}
{{end -}}
{{- if .Install.WantedBy}}
WantedBy={{.Install.WantedBy}}
{{end -}}
{{- if .Install.RequiredBy}}
RequiredBy={{.Install.RequiredBy}}
®{{end -}}
`
)

// WriteUnitToDisk writes a service unit file to disk.
func WriteUnitToDisk(componentName, unitsDir string, unit UnitSpec) error {
	if err := os.MkdirAll(unitsDir, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directory %q", unitsDir)
	}

	tpl := template.Must(template.New("unit").Parse(unitTemplate))
	var buf bytes.Buffer
	err := tpl.Execute(&buf, unit)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal unit for %q", componentName)
	}

	filename := kubeadmconstants.GetSystemUnitFilepath(componentName, unitsDir)

	if err := ioutil.WriteFile(filename, buf.Bytes(), 0600); err != nil {
		return errors.Wrapf(err, "failed to write unit file for %q (%q)", componentName, filename)
	}

	return nil
}
