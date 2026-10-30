// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSampleConfigParses(t *testing.T) {
	// The sample's placeholder secret must be replaced before the server starts.
	sample := filepath.Join("..", "config.yml.sample")
	if err := SetupAppConfig(sample); err == nil || !strings.Contains(err.Error(), "hashing-secret") {
		t.Fatalf("sample config with placeholder secret: err = %v", err)
	}
	raw, _ := os.ReadFile(sample)
	configured := filepath.Join(t.TempDir(), "300sf.yml")
	withHashing := strings.Replace(string(raw), `"change-me"`, `"0123456789abcdef0123456789abcdef"`, 1)
	os.WriteFile(configured, []byte(withHashing), 0o600)
	if err := SetupAppConfig(configured); err == nil || !strings.Contains(err.Error(), "token-secret") {
		t.Fatalf("sample config with placeholder token secret: err = %v", err)
	}
	os.WriteFile(configured, []byte(strings.Replace(withHashing, `"change-me"`, `"fedcba9876543210fedcba9876543210"`, 1)), 0o600)
	if err := SetupAppConfig(configured); err != nil {
		t.Fatal(err)
	}
	if AppConfig.Address != "127.0.0.1:8090" {
		t.Errorf("address = %q", AppConfig.Address)
	}
	if AppConfig.AccessTokenLifetime != 3600 {
		t.Errorf("access-token-lifetime = %d", AppConfig.AccessTokenLifetime)
	}
	if AppConfig.DefaultLists == nil {
		t.Error("default-lists should parse to an empty list")
	}
}

func TestValidateTokenSecret(t *testing.T) {
	hashing := "0123456789abcdef0123456789abcdef"
	for secret, ok := range map[string]bool{
		"":                                 false,
		"change-me":                        false,
		"short":                            false,
		hashing:                            false, // can't be rotated without breaking API keys
		"fedcba9876543210fedcba9876543210": true,
	} {
		err := (&App{HashingSecret: hashing, TokenSecret: secret}).Validate()
		if (err == nil) != ok {
			t.Errorf("token-secret %q: err = %v", secret, err)
		}
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
