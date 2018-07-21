package creds

import (
	"bufio"
	"fmt"
	"forjj/sources_info"
	"io"
	"io/ioutil"
	"os"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"
)

type yamlSecure struct {
	file       string
	credFile   string
	files      []string
	fileToLoad string
	secretFile bool
	file_path  string
	loaded     bool
	Version    string
	Forj       map[string]string
	Objects    map[string]map[string]map[string]*goforjj.ValueStruct
	sources    *sourcesinfo.Sources
	s          *Secrets
}

func (d *yamlSecure) isLoaded() bool {
	if d == nil {
		return false
	}
	return d.loaded
}

func (d *yamlSecure) foundFiles() (ret []string) {
	d.files = make([]string, 2)

	for i, file := range []string{d.credFile, d.file} {
		_, err := os.Stat(file)
		if err == nil {
			d.files[i] = file
		}
	}
	return d.files
}

// load Read the file from encrypted or not file type
func (d *yamlSecure) load(env string, secretFile bool) error {
	d.foundFiles()
	file := d.files[0]
	if ! secretFile {
		file = d.files[1]
	}
	if file == "" {
		gotrace.Trace("no env '%s' file loaded. Not found.", env)
		d.files[0] = d.credFile
		return nil
	}
	fd, err := os.Open(file)
	if err != nil {
		return err
	}

	defer fd.Close()
	if secretFile {
		var data []byte
		data, err = ioutil.ReadAll(fd)
		if err != nil {
			return fmt.Errorf("Unable to read '%s'. %s", d.fileToLoad, err)
		}
		err = d.s.ImportToEnv(data, d)
		if err != nil {
			return fmt.Errorf("Unable to decrypt '%s'. %s", d.fileToLoad, err)
		}
	} else {
		d.iLoad(bufio.NewReader(fd))
	}

	gotrace.Trace("Credential file '%s' has been loaded.", file)
	return nil
}

func (d *yamlSecure) iLoad(r io.Reader) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(d)
}

func (d *yamlSecure) save(secretFile bool) (err error) {
	var (
		yamlData []byte
	)
	file := d.credFile
	if ! secretFile {
		file = d.file
		yamlData, err = yaml.Marshal(d)
	} else {
		yamlData, err = d.s.ExportEnv(d)
	}

	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(d.credFile, yamlData, 0644); err != nil {
		return err
	}
	gotrace.Trace("File name saved: %s", file)
	return
}

func (d *yamlSecure) SetForjValue(source, key, value string) (updated bool) {

	d.sources = d.sources.Set(source, key, value)
	if d.Forj == nil {
		d.Forj = make(map[string]string)
	}
	if v, found := d.Forj[key]; !found || v != value {
		d.Forj[key] = value
		updated = true
	}
	return
}

func (d *yamlSecure) GetForjValue(key string) (ret string, found bool) {
	ret, found = d.Forj[key]
	return
}

func (d *yamlSecure) unsetObjectValue(source, obj_name, instance_name, key_name string) (updated bool) {
	if d.Objects == nil {
		return
	}
	if i, found := d.Objects[obj_name]; !found {
		return
	} else if k, found := i[instance_name]; !found {
		return
	} else if _, found := k[key_name]; !found {
		return
	} else {
		delete(k, key_name)
		i[instance_name] = k
		d.Objects[obj_name] = i
		updated = true
	}
	return
}

func (d *yamlSecure) setObjectValue(source, obj_name, instance_name, key_name string, value *goforjj.ValueStruct) (updated bool) {
	if d.Objects == nil {
		d.Objects = make(map[string]map[string]map[string]*goforjj.ValueStruct)
	}
	var instances map[string]map[string]*goforjj.ValueStruct
	var keys map[string]*goforjj.ValueStruct
	if i, found := d.Objects[obj_name]; !found {
		keys = make(map[string]*goforjj.ValueStruct)
		instances = make(map[string]map[string]*goforjj.ValueStruct)
		newValue := new(goforjj.ValueStruct)
		*newValue = *value
		keys[key_name] = newValue
		instances[instance_name] = keys
		d.Objects[obj_name] = instances
		updated = true
	} else if k, found := i[instance_name]; !found {
		keys = make(map[string]*goforjj.ValueStruct)
		newValue := new(goforjj.ValueStruct)
		*newValue = *value
		keys[key_name] = newValue
		d.Objects[obj_name][instance_name] = keys
		updated = true
	} else if v, found := k[key_name]; found {
		if !value.Equal(v) {
			*v = *value
			updated = true
		}
	} else {
		newValue := new(goforjj.ValueStruct)
		*newValue = *value
		k[key_name] = newValue
		updated = true
	}
	d.sources = d.sources.Set(source, obj_name+"/"+instance_name+"/"+key_name, value.GetString())
	return
}

func (d *yamlSecure) getString(obj_name, instance_name, key_name string) (string, bool, string) {
	v, found, source := d.get(obj_name, instance_name, key_name)
	return v.GetString(), found, source
}

func (d *yamlSecure) get(obj_name, instance_name, key_name string) (ret *goforjj.ValueStruct, found bool, source string) {
	if i, isFound := d.Objects[obj_name]; isFound {
		if k, isFound := i[instance_name]; isFound {
			if v, isFound := k[key_name]; isFound && v != nil {
				ret = new(goforjj.ValueStruct)
				*ret = *v
				found = true
				source = d.sources.Get(obj_name + "/" + instance_name + "/" + key_name)
				return
			}
		}
	}
	return
}

func (d *yamlSecure) getObjectInstance(obj_name, instance_name string) map[string]*goforjj.ValueStruct {
	if i, found := d.Objects[obj_name]; found {
		if k, found := i[instance_name]; found {
			return k
		}
	}
	return nil
}
