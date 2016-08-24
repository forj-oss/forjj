package main

import (
    "path"
    "time"
    "fmt"
)

const (
    defaultTimeout = 32 * time.Second
    default_socket_baseurl = "http:///anyhost"
    default_mount_path = "/src"
)

// Start driver task.
// Forj.CurrentPluginDriver is set to the current driver
func (a *Forj) driver_do(instance_name, action string, args ...string) (err error, aborted bool) {
    d := a.drivers[instance_name]
    a.CurrentPluginDriver = d

    if err := d.plugin.PluginInit(a.Workspace) ; err != nil {
        return err, false
    }

    d.plugin.PluginSetSource(path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.driver_type))
    d.plugin.PluginSetWorkspace(path.Join(a.Workspace_path, a.Workspace))
    d.plugin.PluginSocketPath(path.Join(a.Workspace_path, a.Workspace, "lib"))

    if err := d.plugin.PluginStartService(instance_name) ; err != nil {
        return err, false
    }

    plugin_args := make(map[string]string)
    a.drivers_options.GetDriversMaintainParameters(plugin_args, action)
    a.GetDriversActionsParameters(plugin_args, "common")
    a.GetDriversActionsParameters(plugin_args, action)

    d.plugin.Result, err = d.plugin.PluginRunAction(action, plugin_args)
    if err != nil {
        if d.plugin.Result.State_code == 419 { // The plugin won't do the task because of requirement not met. This is not an error which requires Forjj to exit.
            aborted = true // So, when a plugin return 419, the plugin task is considered as aborted. So forjj can continue if it is possible. (create/update action case)
        }
        return err, aborted
    }

    fmt.Printf("%s:\n%s\n", d.name, d.plugin.Result.Data.Status)

    // store plugins options required at maintain phase from what the plugin returned.
    if action != "maintain" {
        a.drivers_options.AddForjjPluginOptions(d.name, d.plugin.Result.Data.Options, d.driver_type)
    }
    return
}

func (a *Forj) driver_cleanup(driver_type string) {
    d := a.drivers[driver_type]
    d.plugin.PluginStopService()
}
