package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "gopkg.in/alecthomas/kingpin.v2"
    "io/ioutil"
    "net/http"
    "os"
    "os/user"
    "regexp"
    "strings"
)

// Load driver options to a Command requested.

// Currently there is no distinction about setting different options for a specific task on the driver.
func (a *Forj) load_driver_options(service_type string) error {
    if err := a.read_driver(service_type); err != nil {
        return err
    }

    if a.drivers[service_type].plugin.Yaml.Name != "" { // if true => Driver Def loaded
        a.init_driver_flags(service_type)
    }

    return nil
}

// Read Driver yaml document
func (a *Forj) read_driver(service_type string) (err error) {
    var (
        yaml_data   []byte
        driver_name = a.drivers[service_type].name
        source      string
    )

    if driver_name == "" {
        return
    }

    if a.ContribRepo_uri.Scheme == "" {
        // File to read locally
        source = fmt.Sprintf("%s/%s/%s/%s.yaml", a.ContribRepo_uri.Path, service_type, driver_name, driver_name)
        if source[:1] == "~" {
            if user, err := user.Current(); err != nil {
                kingpin.Fatalf("Unable to get your user. %s. Consider to replace ~ by $HOME\n", err)
            } else {
                source = string(regexp.MustCompile("^~").ReplaceAll([]byte(source), []byte(user.HomeDir)))
            }
        }
        gotrace.Trace("Load plugin %s file definition at '%s'\n", service_type, source)
        if d, err := ioutil.ReadFile(source); err != nil {
            return fmt.Errorf("Unable to read '%s'. %s\n", source, err)
        } else {
            yaml_data = d
        }

    } else {
        // File to read for an url. Usually, a raw from github.
        source = fmt.Sprintf("%s/%s/%s/%s/%s.yaml", a.ContribRepo_uri, a.Branch, service_type, driver_name, driver_name)
        gotrace.Trace("Load plugin %s file definition at '%s'\n", service_type, source)

        var resp *http.Response
        if resp, err = http.Get(source); err != nil {
            return fmt.Errorf("Unable to read '%s'. %s\n", source, err)
        }
        defer resp.Body.Close()

        var d []byte
        if d, err = ioutil.ReadAll(resp.Body); err != nil {
            return
        }
        if strings.Contains(http.DetectContentType(d), "text/plain") {
            yaml_data = d
        }
    }

    d := a.drivers[service_type] // Copy of the element. Not a reference.
    if err = d.plugin.PluginDefLoad(yaml_data); err != nil {
        return
    }
    a.drivers[service_type] = d

    return
}

// Initialize command drivers with plugin definition loaded from flags (yaml representation).
func (a *Forj) init_driver_flags(service_type string) {
    commands := a.drivers[service_type].plugin.Yaml.Actions

    gotrace.Trace("Setting flags from plugin type '%s' (%s)\n", service_type, a.drivers[service_type].plugin.Yaml.Name)
    for command, def := range commands {
        if _, ok := a.drivers[service_type].cmds[command]; !ok {
            fmt.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.", a.drivers[service_type], command)
        }

        d := a.drivers[service_type]
        for option_name, params := range def.Flags {
            d.cmds[command].flags[option_name] = "" // No value by default. Will be set later after complete parse.
            // drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
            // They are not in the list of possible drivers options from the cli.
            if ok, _ := regexp.MatchString("forjj-.*", option_name); ok {
                continue
            }

            var flag *kingpin.FlagClause
            // Create flag 'option_name' on kingpin cmd or app
            if command == "common" {
                gotrace.Trace("Set Common flag for '%s'\n", option_name)
                flag = a.app.Flag(option_name, params.Help)
                if d.flags == nil {
                    d.flags = make(map[string]*kingpin.FlagClause)
                    d.flagsv = make(map[string]*string)
                }
                d.flags[option_name] = flag
                d.flagsv[option_name] = flag.String()
            } else {
                gotrace.Trace("Set action '%s' flag for '%s'\n", command, option_name)
                opts := a.GetActionOptsFromString(command)
                flag = opts.Cmd.Flag(option_name, params.Help)
                opts.flags[option_name] = flag
                opts.flagsv[option_name] = flag.String()
            }

            if params.Required {
                flag.Required()
            }
        }
    }

}

func (a *Forj) GetDriversFlags(args []string) {
    a.LoadContext(os.Args[1:])

    if err := a.load_driver_options("ci"); err != nil {
        fmt.Printf("Error: %#v\n", err)
        os.Exit(1)
    }

    if err := a.load_driver_options("upstream"); err != nil {
        fmt.Printf("Error: %#v\n", err)
        os.Exit(1)
    }
}
