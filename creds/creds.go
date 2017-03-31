package creds

import (
	"path"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/forj-oss/forjj-modules/trace"
	"fmt"
)

type YamlSecure struct {
	file string
	updated bool
	Forj map[string]string
	Objects map[string]map[string]map[string]string
}

const default_creds_file = "forjj-creds.yml"

func (d *YamlSecure)Load() error {
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

func (d *YamlSecure)Save() error {
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

func (d *YamlSecure)SetPath(aPath string) {
	d.SetFile(path.Join(aPath, default_creds_file))
}

func (d *YamlSecure)SetFile(file_path string) {
	d.file = path.Clean(file_path)
}

func (d *YamlSecure)SetForjValue(key, value string) {
	if d.Forj == nil {
		d.Forj = make(map[string]string)
	}
	if v, found := d.Forj[key] ; !found || v != value {
		d.Forj[key] = value
		d.updated = true
	}
}

func (d *YamlSecure)SetObjectValue(obj_name, instance_name, key_name, value string) {
	if d.Objects == nil {
		d.Objects = make(map[string]map[string]map[string]string)
	}
	var instances map[string]map[string]string
	var keys map[string]string
	if i, found := d.Objects[obj_name] ; !found  {
		keys = make(map[string]string)
		instances = make(map[string]map[string]string)
		keys[key_name] = value
		instances[instance_name] = keys
		d.Objects[obj_name] = instances
		d.updated = true
	} else {
		if k, found := i[instance_name] ; !found {
			keys = make(map[string]string)
			keys[key_name] = value
			d.Objects[obj_name][instance_name] = keys
			d.updated = true
		} else {
			if v, found := k[key_name] ; !found || value != v {
				k[key_name] = value
				d.updated = true
			}
		}
	}
}

func (d *YamlSecure)Get(obj_name, instance_name, key_name string) (string, bool) {
	if i, found := d.Objects[obj_name] ; found {
		if k, found := i[instance_name] ; found {
			if v, found := k[key_name] ; found {
				return v, true
			}
		}
	}
	return "", false
}

