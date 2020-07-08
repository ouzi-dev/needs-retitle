package config

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/ouzi-dev/needs-retitle/pkg/plugin"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

// ConfigAgent contains the agent mutex and the Agent configuration.
type PluginConfigAgent struct {
	configuration *Configuration
	plugin        *plugin.Plugin
}

// Configuration is the top-level serialization target for plugin Configuration.
type Configuration struct {
	NeedsRetitle NeedsRetitle `json:"needs_retitle"`
}

type NeedsRetitle struct {
	Regexp       string `json:"regexp"`
	ErrorMessage string `json:"error_message"`
}

func NewPluginConfigAgent() *PluginConfigAgent {
	return &PluginConfigAgent{
		plugin: &plugin.Plugin{},
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
	np := &Configuration{}
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
func (pca *PluginConfigAgent) Set(pc *Configuration) {
	pca.configuration = pc
	r, _ := regexp.Compile(pca.configuration.NeedsRetitle.Regexp)
	pca.plugin.SetConfig(pc.NeedsRetitle.ErrorMessage, r)
}

func (c *Configuration) Validate() error {
	if len(c.NeedsRetitle.Regexp) == 0 {
		return fmt.Errorf("needs_pr_rename.regexp can not be empty")
	}

	_, err := regexp.Compile(c.NeedsRetitle.Regexp)

	if err != nil {
		return fmt.Errorf("error compiling regular expression %s: %v", c.NeedsRetitle.Regexp, err)
	}

	return nil
}
