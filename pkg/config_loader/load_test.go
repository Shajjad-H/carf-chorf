package config_loader

import (
	"os"
	"path/filepath"
	"testing"
)

type testYAMLConfig struct {
	Name  string `yaml:"name"`
	Count uint   `yaml:"count"`
}

func TestLoadYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("name: test\ncount: 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg testYAMLConfig
	if err := loadYAML(path, &cfg); err != nil {
		t.Fatalf("loadYAML() error = %v", err)
	}
	if cfg.Name != "test" || cfg.Count != 2 {
		t.Fatalf("loadYAML() = %+v", cfg)
	}
}

func TestLoadYAMLRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("name: test\nunknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg testYAMLConfig
	if err := loadYAML(path, &cfg); err == nil {
		t.Fatal("loadYAML() unexpectedly accepted an unknown field")
	}
}

func TestLoadYAMLEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg testYAMLConfig
	if err := loadYAML(path, &cfg); err == nil {
		t.Fatal("loadYAML() unexpectedly accepted an empty file")
	}
}

func TestLoadYAMLMissingFile(t *testing.T) {
	var cfg testYAMLConfig
	if err := loadYAML(filepath.Join(t.TempDir(), "missing.yml"), &cfg); err == nil {
		t.Fatal("loadYAML() unexpectedly accepted a missing file")
	}
}
