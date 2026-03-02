# configuration

> Environment configuration with caarlos0/env and godotenv

## app/shared/configuration/conf.go

```go
package configuration

import (
	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewConf)

type Conf struct {
	PORT         string `env:"PORT" envDefault:"8080"`
	PROJECT_NAME string `env:"PROJECT_NAME"`
	VERSION      string `env:"VERSION"`

	// --- PostgreSQL Configuration (Optional) ---
	// default to local postgres if not provided by env, excellent for rapid prototyping
	DATABASE_URL string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"`

	// --- EventBroker Factory ---
	// nats | gcp
	EVENT_BROKER string `env:"EVENT_BROKER" envDefault:"nats"`

	// --- GCP Pub/Sub Configuration (Optional) ---
	GOOGLE_PROJECT_ID string `env:"GOOGLE_PROJECT_ID"`
}

// NewConf loads the configuration and provides it.
// It is returned by value because it's lightweight and immutable.
func NewConf() (Conf, error) {
	return Parse[Conf]()
}
```

---

## app/shared/configuration/parse.go

```go
package configuration

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var once sync.Once

// handleEnvLoad logs the result of loading the .env file. Extracted for testability.
func handleEnvLoad(err error) {
	if err != nil {
		slog.Warn(".env not found, loading environment variables from system.")
	} else {
		slog.Info("Environment variables loaded from .env file.")
	}
}

// loadEnvOnce ensures that the .env file is only loaded once per application lifecycle.
func loadEnvOnce() {
	once.Do(func() {
		handleEnvLoad(godotenv.Load())
	})
}

// Parse loads the .env file (if present) and parses the environment variables into the generic struct T.
// Struct T can use `env:"VAR_NAME"` and `envDefault:"default_value"` tags.
func Parse[T any]() (T, error) {
	loadEnvOnce()
	var conf T
	if err := env.Parse(&conf); err != nil {
		return conf, fmt.Errorf("failed to parse configuration: %w", err)
	}
	return conf, nil
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/configuration/conf_test.go

```go
package configuration

import (
	"os"
	"strings"
	"testing"

	"archetype"
)

func TestNewConf_DefaultValues(t *testing.T) {
	// Let's act like main.go and inject the embedded version
	os.Setenv("VERSION", strings.TrimSpace(archetype.Version))
	os.Unsetenv("PORT")
	os.Unsetenv("PROJECT_NAME")

	conf, err := NewConf()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conf.PORT != "8080" {
		t.Errorf("expected default port 8080, got %s", conf.PORT)
	}
	if conf.PROJECT_NAME != "" {
		t.Errorf("expected empty project name, got %s", conf.PROJECT_NAME)
	}
	if conf.VERSION != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", conf.VERSION)
	}
}

func TestNewConf_CustomEnvs(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("PROJECT_NAME", "mytest")
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
	if conf.PROJECT_NAME != "mytest" {
		t.Errorf("expected project name mytest, got %s", conf.PROJECT_NAME)
	}
	if conf.VERSION != "2.0" {
		t.Errorf("expected version 2.0, got %s", conf.VERSION)
	}
}
func TestParse_Error(t *testing.T) {
	os.Setenv("BAD_INT", "not_a_number")
	defer os.Unsetenv("BAD_INT")

	type BadStruct struct {
		Number int `env:"BAD_INT"`
	}
	_, err := Parse[BadStruct]()
	if err == nil {
		t.Error("expected error parsing non-numeric value into int, got nil")
	}
}
```

---

### app/shared/configuration/parse_test.go

```go
package configuration

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleEnvLoad(t *testing.T) {
	t.Run("logs info when no error", func(t *testing.T) {
		handleEnvLoad(nil)
	})

	t.Run("logs warn when error", func(t *testing.T) {
		handleEnvLoad(errors.New("file not found"))
	})
}

func TestParse(t *testing.T) {
	t.Run("parses configuration successfully", func(t *testing.T) {
		type Config struct {
			Port int    `env:"TEST_PARSE_PORT" envDefault:"8080"`
			Host string `env:"TEST_PARSE_HOST" envDefault:"localhost"`
		}

		// Use defaults
		conf, err := Parse[Config]()
		require.NoError(t, err)
		assert.Equal(t, 8080, conf.Port)
		assert.Equal(t, "localhost", conf.Host)

		// Override with env vars
		t.Setenv("TEST_PARSE_PORT", "3000")
		t.Setenv("TEST_PARSE_HOST", "0.0.0.0")
		conf, err = Parse[Config]()
		require.NoError(t, err)
		assert.Equal(t, 3000, conf.Port)
		assert.Equal(t, "0.0.0.0", conf.Host)
	})

	t.Run("returns error when parsing fails", func(t *testing.T) {
		type BadConfig struct {
			Port int `env:"TEST_PARSE_BAD_PORT" envDefault:"8080"`
		}

		t.Setenv("TEST_PARSE_BAD_PORT", "not-a-number")
		_, err := Parse[BadConfig]()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse configuration")
	})

	t.Run("returns error when required field is missing", func(t *testing.T) {
		type RequiredConfig struct {
			APIKey string `env:"TEST_PARSE_REQUIRED_API_KEY,required"`
		}

		os.Unsetenv("TEST_PARSE_REQUIRED_API_KEY")
		_, err := Parse[RequiredConfig]()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse configuration")
	})
}
```
