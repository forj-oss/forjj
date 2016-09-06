package main


import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "fmt"
    "path"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "os"
)

const (
    forjj_options_file = "forjj-options.yml"
)

// This data structure is going to be saved in the infra repository anytime a global update is done.
type ForjjOptions struct {
    Defaults DefaultsStruct
    Drivers map[string]*Driver
}

type DefaultsStruct struct {
    Flow string
    Instance string `yaml:"upstream-instance"`
}

// Initialize Forjj options
// At least, the infra repo must exists.
func (o *ForjjOptions)Init() {
    if o.Drivers == nil {
        o.Drivers = make(map[string]*Driver)
    }
}

func (o *ForjjOptions)SaveForjjOptions(CommitMsg string) error {
    if err := o.Save(forjj_options_file) ; err != nil {
        return fmt.Errorf("Unable to write '%s'. %s", forjj_options_file, err)
    }

    git("add", forjj_options_file)

    if err := git_commit(CommitMsg, false) ; err != nil {
        return fmt.Errorf("Unable to commit the organization update. %s", err)
    }

    return nil
}

func (f *ForjjOptions)Save(file string) error {
    yaml_data, err := yaml.Marshal(f)
    if err != nil {
        return err
    }

    if err := ioutil.WriteFile(file, yaml_data , 0644 ) ; err != nil {
        return err
    }
    gotrace.Trace("File name saved: %s", file)
    return nil
}

// This functions loads the forjj options definitions from the LoadContext().
func (a *Forj)LoadForjjOptions() error {
    // Read definition file from repo.
    file := path.Clean(path.Join(a.Workspace_path, a.Workspace, a.w.Infra.Name, forjj_options_file))

    if _, err := os.Stat(file) ; err != nil {
        return nil // Nothing to read.
    }

    return a.o.LoadFile(file)
}

func (fp *ForjjOptions)LoadFile(file string) error {
    yaml_data, err := ioutil.ReadFile(file)
    if err != nil {
        return fmt.Errorf("Unable to read '%s'. %s", drivers_data_options_file, err)
    }

    if err := yaml.Unmarshal(yaml_data, fp) ; err != nil {
        return fmt.Errorf("Unable to decode the required plugin options from yaml format for maintain phase. %s.", err)
    }
    return nil
}

func (fo *ForjjOptions)update_options() {
    return
}
