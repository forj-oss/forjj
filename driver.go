package main

import (
    "path"
    "time"
    "log"
    "fmt"
)

const (
    defaultTimeout = 32 * time.Second
    default_socket_baseurl = "http:///anyhost"
    default_mount_path = "/src"
)

// Define starting on this driver
// Forj.CurrentPluginDriver set
func (a *Forj) driver_start(instance string) (error) {

    d, found := a.drivers[instance]
    if !found {
        return fmt.Errorf("Internal error: Unable to find %s from drivers.", instance)
    }
    a.CurrentPluginDriver = d
    return nil
}

// Start driver task.
// Forj.CurrentPluginDriver is set to the current driver
func (d *Driver) driver_do(a *Forj, instance_name, action string, args ...string) (err error, aborted bool) {
    log.Print("-------------------------------------------")
    log.Printf("Running %s on %s...", action, instance_name)

    if err := d.plugin.PluginInit(a.Workspace) ; err != nil {
        return err, false
    }

    d.plugin.PluginSetSource(path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.DriverType))
    d.plugin.PluginSetWorkspace(path.Join(a.Workspace_path, a.Workspace))
    d.plugin.PluginSocketPath(path.Join(a.Workspace_path, a.Workspace, "lib"))
    if err := d.plugin.PluginDockerBin(*a.Actions[action].flagsv["docker-exe-path"]) ; err != nil {
        return err, false
    }

    if err := d.plugin.PluginStartService(a.w.Organization + "_" + instance_name) ; err != nil {
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

    // store plugins options required at maintain phase from what the plugin returned.
    if action != "maintain" {
        a.drivers_options.AddForjjPluginOptions(d.Name, d.plugin.Result.Data.Options, d.DriverType)
    }
    return
}

func (a *Forj) driver_cleanup(instance_name string) {
    if d, ok := a.drivers[instance_name] ; ok {
        d.plugin.PluginStopService()
    }
}
