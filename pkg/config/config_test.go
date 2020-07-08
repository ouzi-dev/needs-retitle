package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	pca := NewPluginConfigAgent()

	err := pca.Load("test/noconfig.yaml")

	assert.NoError(t, err)

	assert.Nil(t, pca.plugin.GetConfig())

	err = pca.Load("test/wrongconfig.yaml")

	assert.Error(t, err)

	assert.Nil(t, pca.plugin.GetConfig())

	err = pca.Load("test/config.yaml")

	assert.NoError(t, err)

	assert.NotNil(t, pca.plugin.GetConfig())
}
