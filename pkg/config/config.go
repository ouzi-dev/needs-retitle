package config

import (
	"io/ioutil"
	"time"

	"github.com/ouzi-dev/needs-retitle/pkg/plugin"
	"github.com/ouzi-dev/needs-retitle/pkg/types"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

// ConfigAgent contains the agent mutex and the Agent configuration.
type PluginConfigAgent struct {
	plugin *plugin.Plugin
}

func NewPluginConfigAgent(p *plugin.Plugin) *PluginConfigAgent {
	return &PluginConfigAgent{
		plugin: p,
	}
}

// GetPlugin returns the plugin
func (pca *PluginConfigAgent) GetPlugin() *plugin.Plugin {
	return pca.plugin
}

// Start starts polling path for plugin config. If the first attempt fails,
// then start returns the error. Future errors will halt updates but not stop.
func (pca *PluginConfigAgent) Start(path string) error {
	if err := pca.Load(path); err != nil {
		return err
	}
	ticker := time.Tick(1 * time.Minute)
	go func() {
		for range ticker {
			if err := pca.Load(path); err != nil {
				logrus.WithField("path", path).WithError(err).Error("Error loading plugin config.")
			}
		}
	}()
	return nil
}

// Load attempts to load config from the path. It returns an error if either
// the file can't be read or the configuration is invalid.
func (pca *PluginConfigAgent) Load(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	np := &types.Configuration{}
	if err := yaml.Unmarshal(b, np); err != nil {
		return err
	}

	if err := np.Validate(); err != nil {
		return err
	}

	pca.Set(np)
	return nil
}

// Set sets the plugin agent configuration.
func (pca *PluginConfigAgent) Set(pc *types.Configuration) {
	pca.plugin.SetConfig(pc)
}
