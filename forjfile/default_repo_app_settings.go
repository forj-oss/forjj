package forjfile

import (
	"github.com/forj-oss/forjj-modules/trace"
)

// DefaultRepoAppSettingsStruct contains a collection of default options/values to apply to Repoitories.
type DefaultRepoAppSettingsStruct map[string]string

// Get return value from a key.
func (d DefaultRepoAppSettingsStruct) Get(key string) (value string, found bool) {
	if v, found := d[key]; found {
		return v, found
	}
	if key != "upstream" {
		return
	}
	// TODO: Remove obsolete reference to "upstream-instance"
	gotrace.Warning("Forjfile: `forj-settings/default/upstream-instance` is obsolete and will be ignored in the future." +
		" Please use `forj-settings/default-repo-apps/upstream` instead.")
	return "", false
}
