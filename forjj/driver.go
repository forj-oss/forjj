package main

import (
    "github.hpe.com/christophe-larsonneur/goforjj"
    "gopkg.in/alecthomas/kingpin.v2"
    "os"
    "path"
)

func (a *Forj) driver_do(driver_type, action string, args ...string) {
    d := a.drivers[driver_type]

    ddata := goforjj.PluginNew()

    source_path := path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", driver_type)

    _, err := os.Stat(source_path)
    if os.IsNotExist(err) {
        err = os.MkdirAll(source_path, 0755)
        kingpin.FatalIfError(err, "Unable to create '%s' directory structure", source_path)
    }

    docker_opts := []string{"-v", source_path + ":/src/"}
    args = a.GetDriversCommonParameters(args, "common")
    args = a.GetDriversActionsParameters(args, action)
    ddata.PluginRun(driver_type, d.Yaml.Runtime.Image, action, docker_opts, args)
}
