package service

import (
	"os"
	"testing"
)

func TestParseDartStringConsts(t *testing.T) {
	data, err := os.ReadFile("/Users/flynn/Desktop/sourceCode/easy_moni_new/lib/core/constants/app_strings.dart")
	if err != nil {
		t.Skip("easy_moni app_strings not available")
	}
	keys := ParseDartStringConsts(string(data))
	if len(keys) < 250 {
		t.Fatalf("expected ~301 keys, got %d", len(keys))
	}
	if keys["appTitle"] == "" {
		t.Fatal("missing appTitle")
	}
}
