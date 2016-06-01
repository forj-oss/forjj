package main

import (
	"fmt"
        "os"
	"strings"
//	"github.com/pborman/getopt"
//        "github.com/alecthomas/kingpin"
        "gopkg.in/alecthomas/kingpin.v2"
//	"os/exec"
//        "bytes" 
//        "log"
//	"path"
//        "bufio"
//        "time"
//        "io"
//        "syscall"
//        "gopkg.in/yaml.v2"
        "net/http"
        "net/url"
        "io/ioutil"
        "github.com/smallfish/simpleyaml"
//        "reflect"
)

type ForjContext struct {
        Organization string
        Orga_path string
}

// Define the default Docker image to use for running forjj actions task by drivers.
const Docker_image = "docker.hos.hpecorp.net/devops/forjj"

// TODO: Support multiple contrib sources.
// TODO: Add flag for branch name to ensure local git branch is correct.
type AdditionalFlags struct {
  ContribRepo_uri *url.URL // URL to github raw files
  Branch string      // branch name
  flags map[string]*kingpin.FlagClause // list of additional flags loaded.
  drivers map[string]string // List of drivers to use.
}

type AddFlagOptions struct {
  Help string
  Type string
}

func (*AdditionalFlags) value(context *kingpin.ParseContext, f *kingpin.FlagClause) (value string, found bool){
 for _, element := range context.Elements {
     if flag, ok := element.Clause.(*kingpin.FlagClause); ok && flag == f {
        value = *element.Value
        found = true
     }
  }
 return
}

func (f *AdditionalFlags) load_driver_options(action *kingpin.CmdClause, service_type string) (*simpleyaml.Yaml, error) {
 var (
      yaml_data string
      driver_name string = f.drivers[service_type]
      m *simpleyaml.Yaml = &simpleyaml.Yaml{}
      err error
      source string
      flags []interface{}
     )
 if driver_name == "" {
    return m, nil
 }

 if f.ContribRepo_uri.Scheme == "" {
    // File to read locally
    source = fmt.Sprintf("%s/%s/%s/%s.yaml", f.ContribRepo_uri.Path, service_type, driver_name, driver_name)
    if d, err := ioutil.ReadFile(source) ; err != nil {
       return m, err
    } else {
       yaml_data = string(d)
    }
 } else {
    source = fmt.Sprintf("%s/%s/%s/%s/%s.yaml", f.ContribRepo_uri, f.Branch, service_type, driver_name, driver_name)
    if resp, err := http.Get(source) ; err != nil {
       return m, err
    } else {
       defer resp.Body.Close()

       if d, err := ioutil.ReadAll(resp.Body) ; err != nil {
	  return m, err
       } else {
         if strings.Contains(http.DetectContentType(d), "text/plain") {
            yaml_data = string(d)
         }
       }
    }
 }

 m, err = simpleyaml.NewYaml([]byte(yaml_data))
 if err != nil {
    return m, err
 }

 m = m.Get("flags")
 if m.IsEmpty() {
   fmt.Printf("FORJJ: warning! %s/%s - flags not defined or empty in '%s'\n", service_type, driver_name, source)
   return m, nil
 }

 flags, err = m.Array()
 if err != nil {
   fmt.Printf("FORJJ: warning! %s/%s - flags is in invalid format in '%s'. Expect a list of map.\n%s\n", service_type, driver_name, source, err)
   return m, nil
 }
 // flags is map[interface{}]interface{}
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

 for _, flag := range flags {
   m_flag := flag.(map[interface {}]interface {})
   for o, params := range m_flag {
     option_name := o.(string)
     if params == nil {
        f.flags[option_name] = action.Flag(option_name, "")
        continue
     }
     m_params := params.(map[interface {}]interface {})

     help := toString(m_params["help"])

     f.flags[option_name] = action.Flag(option_name, help)

     if toBool(m_params["required"]) {
        f.flags[option_name] = f.flags[option_name].Required()
     }
   }
 }
 //fmt.Printf("Document content: \n%v\n", m)

 return m, nil
}

// This function is dedicated to provide a string systematically.
// If the type value is not string converted, the string returns will 
// still be there, but empty.
func toString(v interface{}) (result string) {
 result, _ = v.(string)
 return
}

func toBool(v interface{}) (result bool) {
 switch v :=v.(type) {
  case string:
    if v == "true" { result=true }
 }
 return
}

func (f *AdditionalFlags) GetDriversFlags(args []string) {

 context, _ := app.ParseContext(args)

 // Identifying appropriate Contribution Repository.
 if value, found := f.value(context, cr_contrib_repo_flag); found {

   if u, err := url.Parse(value); err != nil {
      println(err)
      os.Exit(1)
   } else {
     f.ContribRepo_uri = u
   }
 }


 // Identifying `ci` drivers options
 if value, found := f.value(context, cr_ci_driver_flag); found {
   f.drivers["ci"] = value
 }

 if _, err := f.load_driver_options(context.SelectedCommand, "ci") ; err != nil {
    fmt.Printf("Error: %#v\n", err)
    os.Exit(1)
 }

 // Identifying `git-us` drivers options
 if value, found := f.value(context, cr_gitus_driver_flag); found {
   f.drivers["upstream"] = value
 }
 if _, err := f.load_driver_options(context.SelectedCommand, "upstream") ; err != nil {
    fmt.Printf("Error: %#v\n", err)
    os.Exit(1)
 }
}

func NewAddFlags() *AdditionalFlags {
 u, _ := url.Parse("https://github.hpe.com/forj/forjj-contribs/raw")
 return &AdditionalFlags{
    ContribRepo_uri: u,
    Branch: "master",
    drivers: make(map[string]string),
    flags: make(map[string]*kingpin.FlagClause),
 }
}

func main() {
 app.Version("forjj V0.0.1 (POC)").Author("Christophe Larsonneur <christophe.larsonneur@hpe.com>")

 drivers_flags := NewAddFlags()
 drivers_flags.GetDriversFlags(os.Args[1:])
 switch kingpin.MustParse(app.Parse(os.Args[1:])) {
   case create.FullCommand():
        //context := ForjContext{}
        //context.Create(*cr_orga, *contrib_repo, *ssh_dir)
        fmt.Printf("Contrib-repo: %#v\n", *cr_contrib_repo)


   case update.FullCommand():
        println("update", *up_orga)
        fmt.Printf("Contrib-repo: %#v\n", *up_contrib_repo)

   case maintain.FullCommand():
        println("maintain", *ma_orga)
   }
}
