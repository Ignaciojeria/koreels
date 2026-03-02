package configuration

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	ledger "ledger-service"
)

func TestNewConf_DefaultValues(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	noenv := filepath.Join(filepath.Dir(f), "testdata", "noenv")
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	_ = os.Chdir(noenv)

	os.Setenv("VERSION", strings.TrimSpace(ledger.Version))
	os.Unsetenv("PORT")
	os.Unsetenv("PROJECT_NAME")

	conf, err := NewConf()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.PORT != "8080" {
		t.Errorf("expected default port 8080, got %s", conf.PORT)
	}
}

func TestNewConf_CustomEnvs(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("PROJECT_NAME", "ledger-service")
	os.Setenv("VERSION", "2.0")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("PROJECT_NAME")
		os.Unsetenv("VERSION")
	}()

	conf, err := NewConf()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.PORT != "9090" {
		t.Errorf("expected port 9090, got %s", conf.PORT)
	}
}
