package main

import (
    "github.hpe.com/christophe-larsonneur/goforjj"
    "gopkg.in/alecthomas/kingpin.v2"
    "os"
    "path"
    "fmt"
)

func (a *Forj) driver_do(driver_type, action string, args ...string) {
    d := a.drivers[driver_type]

    switch d.Yaml.Runtime.Service_type {
    case "REST API":
        a.restDriverDo(d, action, args...)
    case "shell":
        a.shellDriverDo(d, action, args...)
    default:
        fmt.Printf("Warning! Invalid '%s' service_type. Supports only 'REST API' and 'shell'. Use shell as default.", d.Yaml.Runtime.Service_type)
        a.shellDriverDo(d, action, args...)
    }
}

func (a *Forj) restDriverDo(d Driver, action string, args ...string) {
    fmt.Printf("Not yet implemented\n")
}

func (a *Forj) shellDriverDo(d Driver, action string, args ...string) {
    source_path := path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.driver_type)

    _, err := os.Stat(source_path)
    if os.IsNotExist(err) {
        err = os.MkdirAll(source_path, 0755)
        kingpin.FatalIfError(err, "Unable to create '%s' directory structure", source_path)
    }

    if d.Yaml.Runtime.Image != "" {
        a.DriverDockerRun(d.driver_type, d.Yaml.Runtime.Image, source_path, action, args...)
    } else {
        a.DriverRun(source_path, action, args...)
    }
}

func (a *Forj)DriverDockerRun(driver_type, image, source_path, action string, args ...string){
    ddata := goforjj.PluginNew()
    docker_opts := []string{"-v", source_path + ":/src/"}
    args = a.GetDriversCommonParameters(args, "common")
    args = a.GetDriversActionsParameters(args, action)
    ddata.PluginDockerRun(driver_type, image, action, docker_opts, args)
}

func (a *Forj)DriverRun(source_path, action string, args ...string) {
    fmt.Printf("Not yet implemented\n")
}
