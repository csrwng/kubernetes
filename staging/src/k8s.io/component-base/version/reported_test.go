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

import (
	"os"
	"path/filepath"
	"testing"
)

// Downstream (OpenShift) carry tests for the KAS_REPORTED_VERSION_FILE
// masking mechanism. See reported.go.

func TestLoadReportedVersionFile(t *testing.T) {
	valid := `{"major":"1","minor":"34","gitVersion":"v1.34.2","gitCommit":"9a1f2c3d4e5b6a7f8c9d0e1f2a3b4c5d6e7f8a9b","gitTreeState":"clean","buildDate":"2026-08-12T10:14:33Z","goVersion":"go1.23.4","compiler":"gc","platform":"linux/amd64"}`

	tests := []struct {
		name        string
		content     string
		expectErr   bool
		expectMajor string
	}{
		{
			name:        "valid file",
			content:     valid,
			expectMajor: "1",
		},
		{
			name:      "malformed JSON",
			content:   `{"major":`,
			expectErr: true,
		},
		{
			name:      "missing required field",
			content:   `{"major":"1","minor":"34"}`,
			expectErr: true,
		},
		{
			name:      "empty required field",
			content:   `{"major":"1","minor":"34","gitVersion":"","gitCommit":"c","gitTreeState":"clean","buildDate":"d","goVersion":"g","compiler":"gc","platform":"linux/amd64"}`,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "version.json")
			if err := os.WriteFile(path, []byte(test.content), 0o644); err != nil {
				t.Fatal(err)
			}
			info, err := LoadReportedVersionFile(path)
			if test.expectErr {
				if err == nil {
					t.Fatalf("expected error, got info %+v", info)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Major != test.expectMajor {
				t.Errorf("expected major %q, got %q", test.expectMajor, info.Major)
			}
		})
	}
}

func TestLoadReportedVersionFileMissing(t *testing.T) {
	if _, err := LoadReportedVersionFile(filepath.Join(t.TempDir(), "does-not-exist.json")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReportedVersionUnset(t *testing.T) {
	// The environment variable is not set in tests, so the package-level
	// override must be nil and stock behavior is preserved.
	if ReportedVersion() != nil {
		t.Errorf("expected no reported version override, got %+v", ReportedVersion())
	}
}
