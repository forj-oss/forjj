package creds

import (
	"fmt"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

const (
	CredsVersion = "0.1"
)

// Secure is the master object to control Forjj security information.
type Secure struct {
	defaultPath string
	curEnv      string
	updated     bool
	envs        map[string]yamlSecure
}

// DefaultCredsFile is the default credential file name, without environment information.
const (
	DefaultCredsFile = "forjj-creds.yml"
	Global           = "global"
)

// Upgrade detects a need to upgrade current credentials data to new version
func (d *Secure) Upgrade(v0Func func(*Secure, string) error) (_ error) {
	if d == nil {
		return fmt.Errorf("Unable to upgrade. nil object.")
	}
	if d.defaultPath == "" {
		return fmt.Errorf("Internal error! creds.InitEnvDefaults not called with correct filename")
	}
	// Identifying V0 - No deployment
	if err := v0Func(d, "V0.1"); err != nil { // Upgrade from V0 to V0.1
		return fmt.Errorf("Unable top upgrade credential data. %s", err)
	}
	return
}

// IsLoaded return true if the env file were loaded. successfully.
func (d *Secure) IsLoaded(env string) (_ bool) {
	if d == nil {
		return
	}

	if v, found := d.envs[env]; found {
		return v.isLoaded()
	}
	return
}

// IsOld return true if the env file loaded is old.
func (d *Secure) Version(env string) (_ string) {
	if d == nil {
		return
	}

	if v, found := d.envs[env]; found && v.isLoaded() {
		if v.Version == "" {
			return "V0"
		}
		return v.Version
	}
	return
}

// DirName Return the directory name owning the security file
func (d *Secure) DirName(env string) (_ string) {
	if v, found := d.envs[env]; found {
		return v.file_path
	}
	return
}

// Load security files (global + deployment one)
func (d *Secure) Load() error {
	inError := false
	for _, env := range d.envs {
		if _, err := os.Stat(env.file); err != nil {
			gotrace.Trace(" '%s'. %s. Ignored", env.file, err)
			continue
		}
		if err := env.load(); err != nil {
			gotrace.Error("%s", err)
			inError = true
		}
	}
	if inError {
		return fmt.Errorf("Issues detected while loading credential files")
	}
	return nil
}

// Save security files (global + deployment one)
func (d *Secure) Save() error {
	inError := false
	for _, env := range d.envs {
		if err := env.save(); err != nil {
			gotrace.Error("%s", err)
			inError = true
		}
	}
	if inError {
		return fmt.Errorf("Issues detected while loading credential files")
	}
	return nil
}

// InitEnvDefaults initialize the internal cred module with file path.
// the file is prefixed by the deployment environment name.
func (d *Secure) InitEnvDefaults(aPath, env string) {
	d.defaultPath = aPath
	d.envs = make(map[string]yamlSecure)
	for _, curEnv := range []string{Global, env} {
		d.SetDefaultFile(curEnv)
	}
	d.curEnv = env
}

// DefineDefaultCredFileName define the internal credential path file for a specific environment.
func (d *Secure) DefineDefaultCredFileName(aPath, env string) string {
	if env == Global {
		return path.Join(aPath, DefaultCredsFile)
	}
	return path.Join(aPath, env+"-"+DefaultCredsFile)
}

// SetDefaultFile
func (d *Secure) SetDefaultFile(env string) {
	if d == nil || d.envs == nil {
		return
	}
	data := yamlSecure{
		Version:   CredsVersion,
		file:      path.Clean(d.DefineDefaultCredFileName(d.defaultPath, env)),
		file_path: d.defaultPath,
	}
	d.envs[env] = data
	return
}

// SetFile load a single file for the env given.
// if env is 'global', so data is considered as valid for all environment.
func (d *Secure) SetFile(filePath, env string) {
	if d == nil || d.envs == nil {
		return
	}
	data := yamlSecure{
		Version:   CredsVersion,
		file:      path.Clean(filePath),
		file_path: path.Dir(filePath),
	}
	d.envs[env] = data
}

// SetForjValue set a value in Forj section.
func (d *Secure) SetForjValue(env, key, value string) (_ bool, _ error) {
	if v, found := d.envs[env]; found {
		if v.SetForjValue(key, value) {
			d.envs[env] = v
			d.updated = true
		}
		return d.updated, nil
	}
	return d.updated, fmt.Errorf("Credential env '%s' nt found", env)
}

// SetForjValue set a value in Forj section.
func (d *Secure) GetForjValue(env, key string) (_ string, _ bool) {
	if v, found := d.envs[env]; found {
		return v.GetForjValue(key)
	}
	return
}

// SetObjectValue set object value
func (d *Secure) SetObjectValue(env, obj_name, instance_name, key_name string, value *goforjj.ValueStruct) (_ bool) {
	if v, found := d.envs[env]; found {
		if v.setObjectValue(obj_name, instance_name, key_name, value) {
			d.updated = true
			d.envs[env] = v
			return true
		}
	}
	return
}

// GetString return a string representation of the value.
func (d *Secure) GetString(objName, instanceName, keyName string) (value string, found bool) {
	for _, env := range []string{d.curEnv, Global} {
		if v, isFound := d.envs[env]; isFound {
			if value, found = v.getString(objName, instanceName, keyName); found {
				return
			}
		}
	}
	return
}

// Get value of the object instance key...
func (d *Secure) Get(objName, instanceName, keyName string) (value *goforjj.ValueStruct, found bool) {
	for _, env := range []string{d.curEnv, Global} {
		if v, isFound := d.envs[env]; isFound {
			if value, found = v.get(objName, instanceName, keyName); found {
				return
			}
		}
	}
	return nil, false
}

// GetObjectInstance return the instance data
func (d *Secure) GetObjectInstance(objName, instanceName string) (values map[string]*goforjj.ValueStruct) {
	if v, found := d.envs[Global]; found {
		values = v.getObjectInstance(objName, instanceName)
		if v, found = d.envs[d.curEnv]; found {
			for name, value := range v.getObjectInstance(objName, instanceName) {
				values[name] = value
			}
		}
	}
	return
}
