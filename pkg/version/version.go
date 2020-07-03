package version

import (
	"fmt"
)

var (
	Version   = "canary"
	GitCommit = ""
)

func GetVersion() string {
	v := fmt.Sprintf("Version: %s", Version)
	if len(GitCommit) > 0 {
		v += fmt.Sprintf("-%s", GitCommit)
	}

	return v
}
