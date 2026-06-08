package config

import "testing"

func TestConfigFilePathDefaultsToDev(t *testing.T) {
	got := ConfigFilePath("")
	want := "configs/config.dev.yaml"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestConfigFilePathUsesEnvironmentName(t *testing.T) {
	got := ConfigFilePath("prod")
	want := "configs/config.prod.yaml"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
