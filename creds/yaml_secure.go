package creds

import (
	"io/ioutil"
	"fmt"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"

)


type yamlSecure struct {
	file      string
	file_path string
	loaded    bool
	Version   string
	Forj      map[string]string
	Objects   map[string]map[string]map[string]*goforjj.ValueStruct
}

func (d *yamlSecure) isLoaded() bool {
	if d == nil {
		return false
	}
	return d.loaded
}

func (d *yamlSecure) load() error {
	yamlData, err := ioutil.ReadFile(d.file)
	if err != nil {
		return fmt.Errorf("Unable to read '%s'. %s", d.file, err)
	}

	if err := yaml.Unmarshal(yamlData, d); err != nil {
		return fmt.Errorf("Unable to load credentials. %s", err)
	}
	gotrace.Trace("Credential file '%s' has been loaded.", d.file)
	return nil
}

func (d *yamlSecure) save() error {
	yamlData, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(d.file, yamlData, 0644); err != nil {
		return err
	}
	gotrace.Trace("File name saved: %s", d.file)
	return nil
}

func (d *yamlSecure) SetForjValue(key, value string) (updated bool) {
	
	if d.Forj == nil {
		d.Forj = make(map[string]string)
	}
	if v, found := d.Forj[key]; !found || v != value {
		d.Forj[key] = value
		updated = true
	}
	return
}

func (d *yamlSecure) setObjectValue(obj_name, instance_name, key_name string, value *goforjj.ValueStruct) (updated bool) {
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
		updated = true
	} else {
		if k, found := i[instance_name]; !found {
			keys = make(map[string]*goforjj.ValueStruct)
			keys[key_name] = value
			d.Objects[obj_name][instance_name] = keys
			updated = true
		} else {
			if v, found := k[key_name]; !found || !value.Equal(v) {
				k[key_name] = value
				updated = true
			}
		}
	}
	return
}

func (d *yamlSecure) getString(obj_name, instance_name, key_name string) (string, bool) {
	v, found := d.get(obj_name, instance_name, key_name)
	return v.GetString(), found
}

func (d *yamlSecure) get(obj_name, instance_name, key_name string) (*goforjj.ValueStruct, bool) {
	if i, found := d.Objects[obj_name]; found {
		if k, found := i[instance_name]; found {
			if v, found := k[key_name]; found {
				return v, true
			}
		}
	}
	return nil, false
}

func (d *yamlSecure) getObjectInstance(obj_name, instance_name string) map[string]*goforjj.ValueStruct {
	if i, found := d.Objects[obj_name]; found {
		if k, found := i[instance_name]; found {
			return k
		}
	}
	return nil
}
