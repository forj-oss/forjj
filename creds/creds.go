package creds

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"
)

const (
	CredsVersion = "0.1"
)

type YamlSecure struct {
	file      string
	file_path string
	Version   string
	updated   bool
	Forj      map[string]string
	Objects   map[string]map[string]map[string]*goforjj.ValueStruct
}

// DefaultCredsFile is the default credential file name, without environment information. 
const DefaultCredsFile = "forjj-creds.yml"

// Upgrade detects a need to upgrade current credentials data to new version
func (d *YamlSecure) Upgrade(v0Func func(*YamlSecure, string) error) (_ error) {
	if d.file == "" {
		return fmt.Errorf("Internal error! creds.SetFile or creds.SetPath not called with correct filename")
	}
	// Identifying V0 - No deployment
	if err := v0Func(d, "V0.1"); err != nil { // Upgrade from V0 to V0.1
		return fmt.Errorf("Unable top upgrade credential data. %s", err)
	}
	return
}

func (d *YamlSecure) DirName() string {
	return d.file_path
}

func (d *YamlSecure) Load() error {
	yaml_data, err := ioutil.ReadFile(d.file)
	if err != nil {
		return fmt.Errorf("Unable to read '%s'. %s", d.file, err)
	}

	if err := yaml.Unmarshal(yaml_data, d); err != nil {
		return fmt.Errorf("Unable to load credentials. %s.", err)
	}
	gotrace.Trace("Credential file '%s' has been loaded.", d.file)
	return nil
}

func (d *YamlSecure) Save() error {
	yaml_data, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(d.file, yaml_data, 0644); err != nil {
		return err
	}
	gotrace.Trace("File name saved: %s", d.file)
	return nil
}

// SetPath initialize the internal cred module with file path.
// the file is prefixed by the deployment environment name.
func (d *YamlSecure) SetPath(aPath, env string) {
	d.SetFile(d.DefineDefaultCredFileName(aPath, env))
}

func (d *YamlSecure) SetFile(file_path string) {
	d.Version = CredsVersion
	d.file = path.Clean(file_path)
	d.file_path = path.Dir(d.file)
}

func (d *YamlSecure) SetForjValue(key, value string) {
	if d.Forj == nil {
		d.Forj = make(map[string]string)
	}
	if v, found := d.Forj[key]; !found || v != value {
		d.Forj[key] = value
		d.updated = true
	}
}

func (d *YamlSecure) SetObjectValue(obj_name, instance_name, key_name string, value *goforjj.ValueStruct) {
	if d.Objects == nil {
		d.Objects = make(map[string]map[string]map[string]*goforjj.ValueStruct)
	}
	var instances map[string]map[string]*goforjj.ValueStruct
	var keys map[string]*goforjj.ValueStruct
	if i, found := d.Objects[obj_name]; !found {
		keys = make(map[string]*goforjj.ValueStruct)
		instances = make(map[string]map[string]*goforjj.ValueStruct)
		keys[key_name] = value
		instances[instance_name] = keys
		d.Objects[obj_name] = instances
		d.updated = true
	} else {
		if k, found := i[instance_name]; !found {
			keys = make(map[string]*goforjj.ValueStruct)
			keys[key_name] = value
			d.Objects[obj_name][instance_name] = keys
			d.updated = true
		} else {
			if v, found := k[key_name]; !found || !value.Equal(v) {
				k[key_name] = value
				d.updated = true
			}
		}
	}
}

func (d *YamlSecure) GetString(obj_name, instance_name, key_name string) (string, bool) {
	v, found := d.Get(obj_name, instance_name, key_name)
	return v.GetString(), found
}

func (d *YamlSecure) Get(obj_name, instance_name, key_name string) (*goforjj.ValueStruct, bool) {
	if i, found := d.Objects[obj_name]; found {
		if k, found := i[instance_name]; found {
			if v, found := k[key_name]; found {
				return v, true
			}
		}
	}
	return nil, false
}

func (d *YamlSecure) GetObjectInstance(obj_name, instance_name string) map[string]*goforjj.ValueStruct {
	if i, found := d.Objects[obj_name]; found {
		if k, found := i[instance_name]; found {
			return k
		}
	}
	return nil
}

// DefineDefaultCredFileName define the internal credential path file for a specific environment.
func (d *YamlSecure) DefineDefaultCredFileName(aPath, env string) string {
	return path.Join(aPath, env+"-"+DefaultCredsFile)
}
