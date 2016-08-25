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
    "text/template"
    "bytes"
)

// Load driver options to a Command requested.

// Currently there is no distinction about setting different options for a specific task on the driver.
func (a *Forj) load_driver_options(instance_name string) error {
    if err := a.read_driver(instance_name); err != nil {
        return err
    }

    if a.drivers[instance_name].plugin.Yaml.Name != "" { // if true => Driver Def loaded
        a.init_driver_flags(instance_name)
    }

    return nil
}

func (d *Driver)Model() (m *DriverModel) {
    m = &DriverModel{
        InstanceName : d.instance_name,
        Name: d.name,
    }
    return
}

// Read Driver yaml document
func (a *Forj) read_driver(instance_name string) (err error) {
    var (
        yaml_data   []byte
        driver_name = a.drivers[instance_name].name
        service_type = a.drivers[instance_name].driver_type
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
        gotrace.Trace("Load plugin %s file definition at '%s'", service_type, source)
        if d, err := ioutil.ReadFile(source); err != nil {
            return fmt.Errorf("Unable to read '%s'. %s\n", source, err)
        } else {
            yaml_data = d
        }

    } else {
        // File to read from an url. Usually, a raw from github.
        source = fmt.Sprintf("%s/%s/%s/%s/%s.yaml", a.ContribRepo_uri, a.Branch, service_type, driver_name, driver_name)
        gotrace.Trace("Load plugin %s file definition at '%s'", service_type, source)

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

    d := a.drivers[instance_name]
    if err = d.plugin.PluginDefLoad(yaml_data); err != nil {
        return
    }
    // Set defaults value for undefined parameters
    var ff string
    if d.plugin.Yaml.CreatedFile == "" {
        ff =  "." + d.instance_name + ".created"
        d.forjj_flag_file = true // Forjj will test the creation success itself, as the driver did not created it automatically.
    } else {
        ff = d.plugin.Yaml.CreatedFile
    }

    // Initialized defaults value from templates
    var doc bytes.Buffer

    if t, err := template.New("plugin").Parse(ff) ; err != nil {
        return fmt.Errorf("Unable to interpret plugin yaml definition. '/created_flag_file' has an invalid template string '%s'. %s", d.plugin.Yaml.CreatedFile, err)
    } else {
        t.Execute(&doc, d.Model())
    }
    d.flag_file = doc.String()
    gotrace.Trace("Created flag file name Set to default for plugin instance '%s' to %s", d.instance_name, d.plugin.Yaml.CreatedFile)

    return
}

// Initialize command drivers flags with plugin definition loaded from plugin yaml file.
func (a *Forj) init_driver_flags(instance_name string) {
    d := a.drivers[instance_name]
    service_type := d.driver_type
    commands := d.plugin.Yaml.Actions

    gotrace.Trace("Setting flags from plugin type '%s' (%s)", service_type, d.plugin.Yaml.Name)
    for command, def := range commands {
        if _, ok := a.drivers[instance_name].cmds[command]; !ok {
            fmt.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.", service_type, command)
        }

        search_re, _ := regexp.Compile("^(.*[_-])?(" + d.plugin.Yaml.Name +")([_-].*)?$")
        for option_name, params := range def.Flags {

            // drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
            // They are not in the list of possible drivers options from the cli.
            if ok, _ := regexp.MatchString("forjj-.*", option_name); ok {
                d.cmds[command].flags[option_name] = DriverCmdOptionFlag{ driver_flag_name: option_name } // No value by default. Will be set later after complete parse.
                continue
            }

            forjj_option_name := SetAppropriateflagName(option_name, instance_name, search_re)
            d.cmds[command].flags[forjj_option_name] = DriverCmdOptionFlag{ driver_flag_name: option_name } // No value by default. Will be set later after complete parse.

            var flag *kingpin.FlagClause
            // Create flag 'option_name' on kingpin cmd or app
            if command == "common" {
                if forjj_option_name != option_name {
                    gotrace.Trace("Set Common flag for '%s(%s)'", forjj_option_name, option_name)
                } else {
                    gotrace.Trace("Set Common flag for '%s'", forjj_option_name)
                }

                flag = a.app.Flag(forjj_option_name, params.Help)
                if d.flags == nil {
                    d.flags = make(map[string]*kingpin.FlagClause)
                    d.flagsv = make(map[string]*string)
                }
                d.flags[forjj_option_name] = flag
                d.flagsv[forjj_option_name] = flag.String()
            } else {
                if forjj_option_name != option_name {
                    gotrace.Trace("Set action '%s' flag for '%s(%s)'", command, forjj_option_name, option_name)
                } else {
                    gotrace.Trace("Set action '%s' flag for '%s'", command, forjj_option_name)
                }
                opts := a.GetActionOptsFromString(command)
                flag = opts.Cmd.Flag(forjj_option_name, params.Help)
                opts.flags[forjj_option_name] = flag
                opts.flagsv[forjj_option_name] = flag.String()
            }

            if params.Required {
                flag.Required()
            }
        }
    }

}

// Define the Forjj flag name for the plugin selected.
// If instanceName is not equal to service_type, then
// any flag without <service_type> in flag name will returned with <instance_name>-<FlagName>
// Any flag having at least .*[_-]<service_type> or <service_type>[_-].*or both, <service_type is replaced by <instance_name>
func SetAppropriateflagName(flag_name, instance_name string, search_re *regexp.Regexp) string {

    res := search_re.FindStringSubmatch(flag_name)
    if res == nil {

        return instance_name + "-" + flag_name
    }

    return search_re.ReplaceAllString(flag_name, "${1}" + instance_name + "${3}")
}

func (a *Forj) GetDriversFlags(args []string) {
    a.LoadContext(os.Args[1:])

    // Loop on drivers to pre-initialized drivers flags.
    gotrace.Trace("Number of plugins provided from parameters: %d", len(a.drivers_list.list))
    for _, d := range a.drivers_list.list {
        a.drivers[d.Instance] = &Driver{
            name:          d.Name,
            driver_type:   d.Type,
            instance_name: d.Instance,
            cmds: map[string]DriverCmdOptions{
                "common":   DriverCmdOptions{make(map[string]DriverCmdOptionFlag)},
                "create":   DriverCmdOptions{make(map[string]DriverCmdOptionFlag)},
                "update":   DriverCmdOptions{make(map[string]DriverCmdOptionFlag)},
                "maintain": DriverCmdOptions{make(map[string]DriverCmdOptionFlag)},
            },
        }
        gotrace.Trace("Selected '%s' driver: %s\n", d.Type, d.Name)

        if err := a.load_driver_options(d.Instance); err != nil {
            fmt.Printf("Error: %#v\n", err)
            os.Exit(1)
        }
    }
}
