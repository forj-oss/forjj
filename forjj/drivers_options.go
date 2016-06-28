package main

import (
        "fmt"
        "os"
        "os/user"
        "strings"
        "net/http"
        "io/ioutil"
        "github.com/smallfish/simpleyaml"
        "regexp"
)


/* Load driver options to a Command requested.

Currently there is no distinction about setting different options for a specific task on the driver.

*/
func (a *Forj) load_driver_options(opts *ActionOpts, service_type string) (err error) {
 var flags map[interface{}]interface{}

 if flags = a.read_driver_description(service_type) ; flags != nil {
    a.init_driver_flags(opts, flags, service_type)
 }
 return
}

/* Read Driver yaml document

*/
func (a *Forj) read_driver_description(service_type string) (flags map[interface{}]interface{}) {
 var (
      yaml_data []byte
      driver_name string = a.drivers[service_type].name
      source string
     )

 if driver_name == "" { return }

 if a.ContribRepo_uri.Scheme == "" {
    // File to read locally
    source = fmt.Sprintf("%s/%s/%s/%s.yaml", a.ContribRepo_uri.Path, service_type, driver_name, driver_name)
    if source[:1] == "~" {
       if user, err := user.Current() ; err != nil {
         fmt.Printf("Unable to get your user. %s\n", err)
       } else {
         source = string(regexp.MustCompile("^~").ReplaceAll([]byte(source), []byte(user.HomeDir)))
       }
    }
    trace("Load plugin %s file definition at '%s'\n", service_type, source)
    if d, err := ioutil.ReadFile(source) ; err != nil {
       fmt.Printf("Unable to read '%s'. %s\n", source, err)
       return
    } else { yaml_data = d }

 } else {
    // File to read for an url. Usually, a raw from github.
    source = fmt.Sprintf("%s/%s/%s/%s/%s.yaml", a.ContribRepo_uri, a.Branch, service_type, driver_name, driver_name)
    trace("Load plugin %s file definition at '%s'\n", service_type, source)
    if resp, err := http.Get(source) ; err != nil {
       fmt.Printf("Unable to read '%s'. %s\n", source, err)
       return

    } else {

       defer resp.Body.Close()

       if d, err := ioutil.ReadAll(resp.Body) ; err != nil {
          return
       } else {
          if strings.Contains(http.DetectContentType(d), "text/plain") { yaml_data = d }
       }
    }
 }

 m, err := simpleyaml.NewYaml([]byte(yaml_data))
 if err != nil {
    fmt.Printf("FORJJ: warning! '%s' is not a valid yaml document. %s\n", source, err)
    return
 }

 m = m.Get("flags")

 flags, err = m.Map()
 if err != nil {
   fmt.Printf("FORJJ: warning! %s/%s - flags is in invalid format in '%s'. Expect a map of typical forjj commands (common/create/update/maintain).\n%s\n", service_type, driver_name, source, err)
 }
 return
}

/* Initialize command drivers with plugin definition loaded from flags (yaml representation).
*/
func (a *Forj) init_driver_flags(opts *ActionOpts, commands_i map[interface{}]interface{}, service_type string) {
 // Small GO explanation:
 //
 // flag is map[interface{}]interface{}
 // So, in a for range loop, key and value are respectively of interface{}
 // If the underlying value is more, like another map of interfaces, we need to assert it.
 // This will dynamically 'cast' the type value to become a map of something.
 // looks like internally, the
 // m_flag := flag.(map[interface {}]interface {})
 // will get the memory address of the underlying value of the interface{} type.
 // So it is not a cast. We are not changing the type from interface{} to map[interface{}]interface{} anymore.

 // warning, if the underlying type is map[interface{}]interface{}. It can be asserted to a
 // map[string]string even if the underlying type of each interface are string...

 // To get the underlying type of an interface value, we can use reflect.TypeOf(v)

 for command_i, i := range commands_i {
   command := command_i.(string)

   if i == nil { continue }

   cmd_opts_i := i.(map[interface{}]interface{})
   for cmd_opt_i, i := range cmd_opts_i {
     cmd_opt := cmd_opt_i.(string)
     if cmd_opt != "flags" { continue }
     flags_i, _ := i.([]interface {})

     if _, ok := a.drivers[service_type].cmds[command] ; !ok {
        fmt.Printf("FORJJ Driver '%s': Invalid tag '%s'. valid one are 'common', 'create', 'update', 'maintain'. Ignored.",
                   a.drivers[service_type], command)
     }
     for _, flag := range flags_i {
       m_flag := flag.(map[interface {}]interface {})
       for o, params := range m_flag {
         option_name := o.(string)

         a.drivers[service_type].cmds[command].flags[option_name] = "" // No value by default. Will be set later after complete parse.
         // drivers flags starting with --forjj are a way to communicate some forjj internal data to the driver.
         // They are not in the list of possible drivers options from the cli.
         if ok, _ := regexp.MatchString("forjj-.*", option_name) ; ok { continue }

         if params == nil {
            flag := opts.Cmd.Flag(option_name, "")
            opts.flags[option_name] = flag
            opts.flagsv[option_name] = flag.String()
            continue
         }
         m_params := params.(map[interface {}]interface {})

         help := to_string(m_params["help"])

         flag := opts.Cmd.Flag(option_name, help)
         opts.flags[option_name] = flag
         opts.flagsv[option_name] = flag.String()

         if to_bool(m_params["required"]) {
            flag.Required()
         }
       }
     }
   }
 }

}

func (a *Forj) GetDriversFlags(args []string) {
 opts := a.LoadContext(os.Args[1:])
 if opts == nil { return }

 if err := a.load_driver_options(opts, "ci") ; err != nil {
    fmt.Printf("Error: %#v\n", err)
    os.Exit(1)
 }

 if err := a.load_driver_options(opts, "upstream") ; err != nil {
    fmt.Printf("Error: %#v\n", err)
    os.Exit(1)
 }
}

