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

    if err := d.plugin.PluginInit(a.Workspace) ; err != nil {
        return nil, err
    }

    d.plugin.PluginSetSource(path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.driver_type))
    d.plugin.PluginSocketPath(path.Join(a.Workspace_path, a.Workspace, "lib"))

    if err := d.plugin.PluginStartService() ; err != nil {
        return nil, err
    }

    return a.driverRunAction(&d, action)
}

func (a *Forj)driverRunAction(d *Driver, action string) (*goforjj.PluginResult, error) {
    // TODO: Must return a map[string]interface{} instead of []string
    args := make(map[string]string)
    a.GetDriversCommonParameters(args)
    a.GetDriversActionsParameters(args, action)
    return d.plugin.PluginRunAction(action, args)
}
