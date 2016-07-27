package main

import (
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

/*
// Do REST Call to the daemon
func (a *Forj)driverRESTGetData(socket, action string, args []string, result *goforjj.PluginResult) {
    req := gorequest.New()
    req.Transport.Dial = func (_, _ string) (net.Conn, error) {
        return net.DialTimeout("unix", socket, defaultTimeout)
    }
    for {
        _, err := os.Stat(socket)
        if os.IsNotExist() {
            time.Sleep(time.Second)
            continue
        }

        var resp gorequest.Response
        var body string

        resp, body, err = req.Get(default_socket_baseurl + "/state").End()
        if body != "READY" {
            time.Sleep(time.Second)
            continue
        }
    }


    req.Get(default_socket_baseurl + "http:///github.sock/ping").End()

}


// Define the list of valid service type.
// invalid one are returned as shell and a warning is returned.
func driver_get_valid_service_type(service_type string) (string, err) {
    switch service_type {
    case "REST API":
        a.restDriverDo(d, action, args...)
    case "shell", "":
        a.shellDriverDo(d, action, args...)
        service_type = "shell"
    default:
        return "", fmt.Errorf("Error! Invalid '%s' service_type. Supports only 'REST API' and 'shell'. Use shell as default.", service_type)
    }
    return service_type, nil
}

// Build the driver arguments
func (a *Forj)get_driver_args(service_type string) (args []string) {
    switch service_type {
    case "REST API":
        a.restDriverDo(d, action, args...)
    case "shell":
        args = a.GetDriversCommonParameters([]string{}, "common")
        args = a.GetDriversActionsParameters(args, action)
        a.shellDriverDo(d, action, args...)
    }
}

// Function which configure docker options (volume, etc...)
// Adapt list of docker options from the service_type.
func (a *Forj) get_docker_opts(service_type, source_path string) (docker_opts []string) {
    switch service_type {
    case "REST API":
        docker_opts = []string{"-v", source_path + ":" + default_mount_path}
        docker_opts = append(docker_opts, "-d") // Daemon mode
        sockets_dir := path.Join(a.Workspace_path, a.Workspace, "lib")
        if _, err := os.Stat(sockets_dir); err == os.IsNotExist() {
            err = os.Mkdir(sockets_dir, 0755)
            kingpin.FatalIfError("Unable to create missing '%s' directory.", sockets_dir)
        }
        docker_opts = append(docker_opts, "-v", sockets_dir + ":/sockets")

    case "shell":
        docker_opts = []string{"-v", source_path + ":" + default_mount_path}
        docker_opts = append(docker_opts, "-i", "--rm") // interactive and removed by default
    }
 return
}

// Run driver as REST API service
func (a *Forj) restDriverDo(d Driver, action string, args ...string) {
    source_path := path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.driver_type)
    if d.Yaml.Runtime.Image != "" {
        docker_opts := []string{"-v", source_path + ":/src/"}
        goforjj.PluginDockerRun(driver_type, image, action, docker_opts, args)
        //if djson, err := a.DriverDockerRun(d.driver_type, d.Yaml.Runtime.Image, source_path, action, args...)
    } else {
        return a.DriverRun(source_path, action, args...)
    }
    req := gorequest.New()
    req.Transport.Dial = socket_dialer
    req.Get("http:///github.sock/ping").End()
}

// Run driver in shell mode (binary and json returned)
func (a *Forj) shellDriverDo(d Driver, action string, args ...string) (*goforjj.PluginResult, error) {
    source_path := path.Join(a.Workspace_path, a.Workspace, a.w.Infra, "apps", d.driver_type)

    _, err := os.Stat(source_path)
    if os.IsNotExist(err) {
        err = os.MkdirAll(source_path, 0755)
        kingpin.FatalIfError(err, "Unable to create '%s' directory structure", source_path)
    }

    if d.Yaml.Runtime.Image != "" {
        return a.DriverDockerRun(d.driver_type, d.Yaml.Runtime.Image, source_path, action, args...)
    } else {
        return a.DriverRun(source_path, action, args...)
    }
}

func driverDockerShellRun(driver_type, image string, docker_opts []string, action string, args ...string) (*goforjj.PluginResult, error) {
    ddata := goforjj.PluginNew()
    if djson, err := goforjj.PluginDockerRun(driver_type, image, action, docker_opts, args); err != nil {
        return nil, fmt.Errorf("Issue to run your driver. %s", err)
    } else {
        json.Unmarshal(djson, &ddata)
    }
    return ddata, nil
}

func (a *Forj)DriverRun(source_path, action string, args ...string) (*goforjj.PluginResult, error) {
    ddata := goforjj.PluginNew()
    return ddata, fmt.Errorf("Not yet implemented\n")
}
*/
