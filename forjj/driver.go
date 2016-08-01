package main

import (
//    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "github.hpe.com/christophe-larsonneur/goforjj"
    "path"
    "time"
)

const (
    defaultTimeout = 32 * time.Second
    default_socket_baseurl = "http:///anyhost"
    default_mount_path = "/src"
)
// Start driver task.
func (a *Forj) driver_do(driver_type, action string, args ...string) (*goforjj.PluginResult, error) {
    d := a.drivers[driver_type]
    a.CurrentPluginDriver = &d

    if err := d.plugin.PluginInit(a.Workspace) ; err != nil {
        return nil, err
    }

    d.plugin.PluginSetSource(path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.driver_type))
    d.plugin.PluginSocketPath(path.Join(a.Workspace_path, a.Workspace, "lib"))

    if err := d.plugin.PluginStartService() ; err != nil {
        return nil, err
    }

    plugin_args := make(map[string]string)
    a.GetDriversActionsParameters(plugin_args, "common")
    a.GetDriversActionsParameters(plugin_args, action)
    return d.plugin.PluginRunAction(action, plugin_args)
}

func (a *Forj) driver_cleanup(driver_type string) {
    d := a.drivers[driver_type]
    d.plugin.PluginStopService()
}
