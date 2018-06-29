package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"log"
	"strings"
)

// DefineDefaultUpstream will set a Defaultupstream if default is not set.
// In that case, if we have no upstream app or 2 or more upstream apps, an error will be returned.
func (a *Forj) DefineDefaultUpstream() error {
	instances := []string{}
	for instance, app := range a.f.Apps() {
		if app.Type == "upstream" {
			instances = append(instances, instance)
		}
	}
	found_instances := len(instances)
	switch {
	case found_instances == 0 :
		return fmt.Errorf("Unable to determine a default upstream instance. No upstream application found.")
	case found_instances >1 :
		return fmt.Errorf("Unable to determine one default upstream instance. " +
			"Found '%s'. You must define one in Forjfile:/settings/default/upstream-instance.",
			strings.Join(instances, "', '"))
	default:
		a.f.Set("forjj", "settings", "default-repo-apps", "upstream", instances[0])
		log.Printf("Set default upstream application to '%s'", instances[0])
	}
	return nil
}

// GetReposRequestedFor Identify number of repository requested for an instance.
func (a *Forj) GetReposRequestedFor(instance, action string) (num int) {
	if instance == "" || action == "" {
		gotrace.Trace("Internal error: instance and action cannot be empty.")
		return
	}
	for _, data := range a.cli.GetObjectValues("repo") {
		if v, _, _ := data.Get("instance"); v == instance || (v == "" && instance == a.o.Defaults["instance"]) {
			num++
		}
	}
	return
}

func NumDisplay(num int, format, elements, element string) string {

	if num > 1 {
		return fmt.Sprintf(format, num, elements)
	}
	return fmt.Sprintf(format, num, element)
}

func NumReposDisplay(num int) string {
	return NumDisplay(num, "%d repositor%s", "ies", "y")
}
