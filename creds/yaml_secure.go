package creds

import (
	"bufio"
	"fmt"
	sourcesinfo "forjj/sources_info"
	"io"
	"io/ioutil"
	"os"
	"sync"

	gotrace "github.com/forj-oss/forjj-modules/trace"
	yaml "gopkg.in/yaml.v2"
)

var data *yamlSecureData

// creds is a module to manage secrets data of forjj.
// It is NOT possible to read secrets recursively, as secret data version is managed through a shared variable
// used at load time (See UnmarshalYAML)
//
// But the module is threadsafe. We can load multiple secrets in parallel but finally will be loaded one by one in series.
// So there is no sense to do it all in parallel.
//
// The version management is made like that, today. If there is a better way to do it, suggest a PR! Until I found a better way to do it.

type yamlSecure struct {
	file       string
	credFile   string
	files      []string
	fileToLoad string
	secretFile bool
	file_path  string
	loaded     bool

	Version string
	Forj    map[string]*Value
	Objects map[string]map[string]map[string]*Value
	sources *sourcesinfo.Sources
	s       *Secrets
}

type yamlSecureData struct {
	Version string
	Forj    map[string]*Value
	Objects map[string]map[string]map[string]*Value
}

func (d *yamlSecure) UnmarshalYAML(unmarchal func(interface{}) error) (err error) {
	mutex := new(sync.Mutex)

	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()

	data = new(yamlSecureData)
	err = unmarchal(data)
	d.Version = data.Version
	if d.Version != CredsVersion {
		gotrace.Trace("Old secret file version loaded: '%s'", d.Version)
	}
	d.Forj = data.Forj
	d.Objects = data.Objects
	data = nil
	return
}

func (d *yamlSecure) isLoaded() bool {
	if d == nil {
		return false
	}
	return d.loaded
}

// initRef set the secret reference to each Value of secrets
func (d *yamlSecure) initRef() {
	if d == nil {
		return
	}

	for _, value := range d.Forj {
		value.setSecrets(d.s)
	}

	for _, instances := range d.Objects {
		for _, values := range instances {
			for _, value := range values {
				value.setSecrets(d.s)
			}
		}
	}
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
	if !secretFile {
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

	// Initialize secrets reference
	d.initRef()
	
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
	d.Version = CredsVersion
	file := d.credFile
	if !secretFile {
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

// setForjValue set a Value in 'forj' section
func (d *yamlSecure) setForjValue(source, key string, value *Value) (updated bool) {
	d.sources = d.sources.Set(source, key, value.value.GetString())
	if d.Forj == nil {
		d.Forj = make(map[string]*Value)
	}
	if v, found := d.Forj[key]; found {
		updated = v.copyFrom(value)
	} else {
		d.Forj[key] = value.clone(d.s)
		updated = true
	}
	return
}

// getForjValue get a value found in 'forj' section
func (d *yamlSecure) getForjValue(key string) (ret *Value, found bool) {
	var value *Value
	if value, found = d.Forj[key]; found {
		ret = NewValue(value.source, value.value)
		ret.resource = value.resource
	}
	return
}

func (d *yamlSecure) unsetObjectValue(obj_name, instance_name, key_name string) (updated bool) {
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

func (d *yamlSecure) setObjectValue(source, obj_name, instance_name, key_name string, value *Value) (updated bool) {
	if d.Objects == nil {
		d.Objects = make(map[string]map[string]map[string]*Value)
	}
	var instances map[string]map[string]*Value
	var keys map[string]*Value
	if i, found := d.Objects[obj_name]; !found {
		keys = make(map[string]*Value)
		instances = make(map[string]map[string]*Value)

		keys[key_name] = value.clone(d.s)
		instances[instance_name] = keys
		d.Objects[obj_name] = instances
		updated = true
	} else if k, found := i[instance_name]; !found {
		keys = make(map[string]*Value)

		keys[key_name] = value.clone(d.s)
		d.Objects[obj_name][instance_name] = keys
		updated = true
	} else if v, found := k[key_name]; !found {
		k[key_name] = value.clone(d.s)
		updated = true
	} else {
		updated = v.copyFrom(value)
	}
	d.sources = d.sources.Set(source, obj_name+"/"+instance_name+"/"+key_name, value.value.GetString())
	return
}

func (d *yamlSecure) getString(obj_name, instance_name, key_name string) (string, bool, string) {
	v, found, source := d.get(obj_name, instance_name, key_name)
	if !found || v == nil || v.value == nil {
		return "", found, source
	}
	return v.value.GetString(), found, source
}

func (d *yamlSecure) get(obj_name, instance_name, key_name string) (ret *Value, found bool, source string) {
	if i, isFound := d.Objects[obj_name]; isFound {
		if k, isFound := i[instance_name]; isFound {
			if v, isFound := k[key_name]; isFound && v != nil && v.value != nil {
				ret = NewValue(v.source, v.value)
				ret.resource = v.resource
				found = true
				source = d.sources.Get(obj_name + "/" + instance_name + "/" + key_name)
				return
			}
		}
	}
	return
}

func (d *yamlSecure) getObjectInstance(obj_name, instance_name string) map[string]*Value {
	if i, found := d.Objects[obj_name]; found {
		if k, found := i[instance_name]; found {
			return k
		}
	}
	return nil
}
