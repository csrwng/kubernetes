/*
Copyright The Kubernetes Authors.

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

package version

// Downstream (OpenShift) carry, not for upstream contribution.
//
// HyperShift control planes may run a control plane payload that differs
// from the payload the hosted cluster admin selected (HostedCluster
// spec.controlPlaneRelease). Every admin-visible version surface follows
// spec.release except the kube-apiserver /version endpoint, which reports
// the compiled-in version of the binary actually running and thereby leaks
// the payload skew.
//
// KAS_REPORTED_VERSION_FILE points at a JSON file (typically ConfigMap
// mounted) carrying the version.Info the kube-apiserver should report on
// /version and in the kubernetes_build_info metric. Upstream will not carry
// a "lie about your version" mechanism on purpose, so this is a permanent
// fork carry. When the variable is unset, behavior is exactly stock.

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"

	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

// ReportedVersionFileEnvVar is the environment variable naming the JSON file
// with the version.Info to report instead of this binary's own.
const ReportedVersionFileEnvVar = "KAS_REPORTED_VERSION_FILE"

// reportedVersion holds the parsed override, or nil when the environment
// variable is unset.
var reportedVersion atomic.Pointer[apimachineryversion.Info]

func init() {
	path := os.Getenv(ReportedVersionFileEnvVar)
	if path == "" {
		return
	}
	info, err := LoadReportedVersionFile(path)
	if err != nil {
		// Fail closed: a configured mask that cannot be used is a startup
		// error, never a silent fallback to the real version.
		panic(fmt.Sprintf("%s: %v", ReportedVersionFileEnvVar, err))
	}
	reportedVersion.Store(info)
}

// ReportedVersion returns the version.Info to report externally, or nil if
// KAS_REPORTED_VERSION_FILE is unset (the default).
func ReportedVersion() *apimachineryversion.Info {
	return reportedVersion.Load()
}

// LoadReportedVersionFile reads and validates a reported-version JSON file.
// Validation is strict: every field exposed by /version and
// kubernetes_build_info must be present, so a partial mask can never leak
// the real build identifiers field by field.
func LoadReportedVersionFile(path string) (*apimachineryversion.Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading reported version file: %w", err)
	}
	var info apimachineryversion.Info
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parsing reported version file %q: %w", path, err)
	}
	fields := map[string]string{
		"major":        info.Major,
		"minor":        info.Minor,
		"gitVersion":   info.GitVersion,
		"gitCommit":    info.GitCommit,
		"gitTreeState": info.GitTreeState,
		"buildDate":    info.BuildDate,
		"goVersion":    info.GoVersion,
		"compiler":     info.Compiler,
		"platform":     info.Platform,
	}
	for name, value := range fields {
		if value == "" {
			return nil, fmt.Errorf("reported version file %q: field %q must not be empty", path, name)
		}
	}
	return &info, nil
}

// SetReportedVersionForTest overrides the reported version; it exists only
// for tests, because the file is loaded during package initialization.
func SetReportedVersionForTest(info *apimachineryversion.Info) {
	reportedVersion.Store(info)
}
