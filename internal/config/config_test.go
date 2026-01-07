package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetConfigFilePath(t *testing.T) {
	expected := "/Users/justin/.gatorconfig.json"
	actual, err := getConfigFilePath()
	if err != nil {
		t.Fatalf("error getting config file path: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestReadConfig(t *testing.T) {
	expected := &Config{
		DbUrl: "postgres://example",
		CurrentUserName: "ShawnSpencer",
	}
	config := NewConfig()
	err := config.Read()
	if err != nil {
		t.Fatalf("error reading config: %v", err)
	}
	diff := cmp.Diff(config, expected)
	if diff != "" {
		t.Fatalf("config mismatch (-want +got):\n%s", diff)
	}
}


func TestSetUserConfig(t *testing.T) {
	expected := &Config{
		DbUrl:            "postgres://example",
		CurrentUserName: "ShawnSpencer",
	}

	config := NewConfig()
	config.DbUrl = expected.DbUrl
	err := config.SetUser(expected.CurrentUserName)
	if err != nil {
		t.Fatalf("error setting user: %v", err)
	}

	err = config.Read()
	if err != nil {
		t.Fatalf("error reading config: %v", err)
	}

	diff := cmp.Diff(expected, config)
	if diff != "" {
		t.Fatalf("config mismatch (-want +got):\n%s", diff)
	}
}
