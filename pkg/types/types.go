package types

import (
	"fmt"
	"regexp"
)

// Configuration is the top-level serialization target for plugin Configuration.
type Configuration struct {
	NeedsRetitle NeedsRetitle `json:"needs_retitle"`
}

// NeedsRetitle is the configuration of the plugin
type NeedsRetitle struct {
	Regexp                     string `json:"regexp"`
	ErrorMessage               string `json:"error_message"`
	RequireEnableAsNotExternal bool   `json:"require_enable_as_not_external"`
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
