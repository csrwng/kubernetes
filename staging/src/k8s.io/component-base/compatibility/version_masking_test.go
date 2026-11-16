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

package compatibility

import (
	"testing"

	apimachineryversion "k8s.io/apimachinery/pkg/version"
	baseversion "k8s.io/component-base/version"
)

// Downstream (OpenShift) carry tests: when KAS_REPORTED_VERSION_FILE is set,
// Info() must report the masked build info (served on /version) while the
// emulation and min-compatibility fields still reflect the real binary.
func TestInfoWithReportedVersion(t *testing.T) {
	masked := &apimachineryversion.Info{
		Major:        "1",
		Minor:        "34",
		GitVersion:   "v1.34.2",
		GitCommit:    "9a1f2c3d4e5b6a7f8c9d0e1f2a3b4c5d6e7f8a9b",
		GitTreeState: "clean",
		BuildDate:    "2026-08-12T10:14:33Z",
		GoVersion:    "go1.23.4",
		Compiler:     "gc",
		Platform:     "linux/amd64",
	}
	baseversion.SetReportedVersionForTest(masked)
	t.Cleanup(func() { baseversion.SetReportedVersionForTest(nil) })

	effective := NewEffectiveVersionFromString("v1.35.0", "", "")
	info := effective.Info()
	if info == nil {
		t.Fatal("expected info, got nil")
	}

	if info.Major != masked.Major || info.Minor != masked.Minor {
		t.Errorf("expected masked major/minor %s/%s, got %s/%s", masked.Major, masked.Minor, info.Major, info.Minor)
	}
	for _, field := range []struct {
		name     string
		got      string
		expected string
	}{
		{"gitVersion", info.GitVersion, masked.GitVersion},
		{"gitCommit", info.GitCommit, masked.GitCommit},
		{"gitTreeState", info.GitTreeState, masked.GitTreeState},
		{"buildDate", info.BuildDate, masked.BuildDate},
		{"goVersion", info.GoVersion, masked.GoVersion},
		{"compiler", info.Compiler, masked.Compiler},
		{"platform", info.Platform, masked.Platform},
	} {
		if field.got != field.expected {
			t.Errorf("%s: expected %q, got %q", field.name, field.expected, field.got)
		}
	}

	// Emulation/min-compatibility fields are derived from the real binary
	// version and must not be masked.
	if info.EmulationMajor != "1" || info.EmulationMinor != "35" {
		t.Errorf("expected real emulation version 1/35, got %s/%s", info.EmulationMajor, info.EmulationMinor)
	}
	if info.MinCompatibilityMajor != "1" || info.MinCompatibilityMinor != "34" {
		t.Errorf("expected real min compatibility version 1/34, got %s/%s", info.MinCompatibilityMajor, info.MinCompatibilityMinor)
	}
}
