package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "net/http"
    "os"
    "os/user"
    "regexp"
    "strings"
    "gopkg.in/alecthomas/kingpin.v2"
)

/* Load driver options to a Command requested.

Currently there is no distinction about setting different options for a specific task on the driver.

*/
func (a *Forj) load_driver_options(opts *ActionOpts, service_type string) (err error) {
    var flags map[string]goforjj.YamlPluginDef

    if flags, err = a.read_driver(service_type) ; err != nil {
        fmt.Printf("%s\n", err)
    } else {
        if a.drivers[service_type].Yaml.Name != "" { // if true => Driver Def loaded
            a.init_driver_flags(opts, flags, service_type)
        }
    }
    return
}

/* Read Driver yaml document

 */
func (a *Forj) read_driver(service_type string) (flags map[string]goforjj.YamlPluginDef, funcerr error) {
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
            funcerr = fmt.Errorf("Unable to read '%s'. %s\n", source, err)
            return
        } else {
            yaml_data = d
        }

    } else {
        // File to read for an url. Usually, a raw from github.
        source = fmt.Sprintf("%s/%s/%s/%s/%s.yaml", a.ContribRepo_uri, a.Branch, service_type, driver_name, driver_name)
        gotrace.Trace("Load plugin %s file definition at '%s'\n", service_type, source)
        if resp, err := http.Get(source); err != nil {
            funcerr = fmt.Errorf("Unable to read '%s'. %s\n", source, err)
            return
        } else {

            defer resp.Body.Close()

            if d, err := ioutil.ReadAll(resp.Body); err != nil {
                funcerr = err
                return
            } else {
                if strings.Contains(http.DetectContentType(d), "text/plain") {
                    yaml_data = d
                }
            }
        }
    }

    d := a.drivers[service_type] // Copy of the element. Not a reference.

    if err := yaml.Unmarshal([]byte(yaml_data), &d.Yaml); err != nil {
        funcerr = fmt.Errorf("FORJJ: warning! '%s' is not a valid yaml document. %s\n", source, err)
        return
    }
    a.drivers[service_type] = d
    flags = d.Yaml.Actions
    return
}

/* Initialize command drivers with plugin definition loaded from flags (yaml representation).
 */
func (a *Forj) init_driver_flags(opts *ActionOpts, commands map[string]goforjj.YamlPluginDef, service_type string) {
    for command, def := range commands {
        if _, ok := a.drivers[service_type].cmds[command]; !ok {
            fmt.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.", a.drivers[service_type], command)
        }
        for option_name, params := range def.Flags {
            a.drivers[service_type].cmds[command].flags[option_name] = "" // No value by default. Will be set later after complete parse.
            // drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
            // They are not in the list of possible drivers options from the cli.
            if ok, _ := regexp.MatchString("forjj-.*", option_name); ok {
                continue
            }

            // Create flag 'option_name' on Cms
            flag := opts.Cmd.Flag(option_name, params.Help)
            opts.flags[option_name] = flag
            opts.flagsv[option_name] = flag.String()

            if params.Required {
                flag.Required()
            }
        }
    }

}

func (a *Forj) GetDriversFlags(args []string) {
    opts := a.LoadContext(os.Args[1:])
    if opts == nil {
        return
    }

    if err := a.load_driver_options(opts, "ci"); err != nil {
        fmt.Printf("Error: %#v\n", err)
        os.Exit(1)
    }

    if err := a.load_driver_options(opts, "upstream"); err != nil {
        fmt.Printf("Error: %#v\n", err)
        os.Exit(1)
    }
}
