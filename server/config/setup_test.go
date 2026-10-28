// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSampleConfigParses(t *testing.T) {
	if err := SetupAppConfig(filepath.Join("..", "config.yml.sample")); err != nil {
		t.Fatal(err)
	}
	if AppConfig.Address != ":8090" {
		t.Errorf("address = %q", AppConfig.Address)
	}
	if AppConfig.AccessTokenLifetime != 3600 {
		t.Errorf("access-token-lifetime = %d", AppConfig.AccessTokenLifetime)
	}
	if AppConfig.DefaultLists == nil {
		t.Error("default-lists should parse to an empty list")
	}
}

func TestSetupAppConfigErrors(t *testing.T) {
	if err := SetupAppConfig(filepath.Join(t.TempDir(), "missing.yml")); err == nil {
		t.Error("expected error for missing file")
	}
	bad := filepath.Join(t.TempDir(), "bad.yml")
	os.WriteFile(bad, []byte("address: [unclosed"), 0o600)
	if err := SetupAppConfig(bad); err == nil {
		t.Error("expected error for invalid yaml")
	}
}

func TestGetLogLevelAsInt(t *testing.T) {
	AppConfig = &App{}
	for level, want := range map[string]int{"silent": 1, "error": 2, "warn": 3, "info": 4, "": 3, "bogus": 3} {
		AppConfig.LogLevel = level
		if got := GetLogLevelAsInt(); got != want {
			t.Errorf("GetLogLevelAsInt(%q) = %d, want %d", level, got, want)
		}
	}
}
