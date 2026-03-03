package observability

import (
	"testing"

	"koreels/internal/shared/configuration"

	"github.com/stretchr/testify/assert"
)

func TestNewObservability(t *testing.T) {
	conf := configuration.Conf{PROJECT_NAME: "koreels", VERSION: "1.0"}
	obs, err := NewObservability(conf)
	assert.NoError(t, err)
	assert.NotNil(t, obs.Tracer)
	assert.NotNil(t, obs.Logger)
}
