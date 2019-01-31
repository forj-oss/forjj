package creds

import (
	"fmt"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
)

const (
	CredsVersion = "0.1"
)

// Secure is the master object to control Forjj security information.
type Secure struct {
	defaultPath string
	curEnv      string
	updated     bool
	key         string
	secrets     Secrets
}

// DefaultCredsFile is the default credential file name, without environment information.
const (
	DefaultCredsFile     = "forjj-creds.yml"
	DefaultSecretFile    = "forjj.enc"
	DefaultSecretKeyFile = ".forjj.key"
	Global               = "global"
)

// Upgrade detects a need to upgrade current credentials data to new version
func (d *Secure) Upgrade(v0Func func(*Secure, string) error) (_ error) {
	if d == nil {
		return fmt.Errorf("Unable to upgrade. nil object")
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

	if v, found := d.secrets.Envs[env]; found {
		return v.isLoaded()
	}
	return
}

// Version return the version of creds loaded.
// If a file is loaded, at least version = V0
// if no file were loaded, verison is empty.
func (d *Secure) Version(env string) (_ string) {
	if d == nil {
		return
	}

	if v, found := d.secrets.Envs[env]; found && v.isLoaded() {
		if v.Version == "" {
			return "V0"
		}
		return v.Version
	}
	return
}

// DirName Return the directory name owning the security file
func (d *Secure) DirName(env string) (_ string) {
	if d == nil {
		return
	}
	if v, found := d.secrets.Envs[env]; found {
		return v.file_path
	}
	return
}

// EncryptAll is executed to encrypt all unencrypted files if found. The process is as follow:
//
// if one uncrypted file is found, and no encrypted found, the file will be encrypted, automatically
// if both unencrypted and encrrypted files are found, it removes the unencrypted file
// if only encrypted file is found, nothing is done
//
// If error is found, the function exit.
func (d *Secure) EncryptAll(encrypt bool) error {
	if _, err := os.Stat(d.key); err != nil && encrypt {
		d.secrets.GenerateKey()
		d.secrets.SaveKey(d.key)
	} else {
		d.secrets.ReadKey(d.key)
	}

	if !encrypt {
		return nil
	}

	for key, env := range d.secrets.Envs {
		files := env.foundFiles()
		if files[0] == "" && files[1] != "" {
			// encrypt the file
			env.load(key, false)
			if err := env.save(true); err != nil {
				return fmt.Errorf("Unable to encrypt '%s'. %s", files[1], err)
			}
			if err := os.Remove(files[1]); err != nil {
				return fmt.Errorf("Unable to remove unencrypted '%s'. %s", files[1], err)
			}
			env.files[0] = env.credFile
			env.files[1] = ""
		} else if files[0] != "" && files[1] != "" {
			// remove unencrypted one
			if err := os.Remove(files[1]); err != nil {
				return fmt.Errorf("Unable to remove unencrypted '%s'. %s", files[1], err)
			}
			env.files[1] = ""
		}
	}
	return nil
}

// Load security files (global + deployment one)
func (d *Secure) Load() error {
	if d == nil {
		return fmt.Errorf("Secure object is nil")
	}
	inError := false

	for key, env := range d.secrets.Envs {
		if err := env.load(key, true); err != nil {
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
	if d == nil {
		return fmt.Errorf("Secure object is nil")
	}
	inError := false
	for _, env := range d.secrets.Envs {
		if err := env.save(true); err != nil {
			gotrace.Error("%s", err)
			inError = true
		}
	}
	if inError {
		return fmt.Errorf("Issues detected while saving credential files")
	}
	return nil
}

// SaveEnv security file.
//
// If env == global, it will save the global file.
func (d *Secure) SaveEnv(env string) error {
	if d == nil {
		return fmt.Errorf("Secure object is nil")
	}
	if envData, found := d.secrets.Envs[env]; found {
		if err := envData.save(true); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unknown deployment environment '%s'", env)
	}
	return nil
}

// InitEnvDefaults initialize the internal cred module with file path.
// the file is prefixed by the deployment environment name.
func (d *Secure) InitEnvDefaults(aPath, env string) {
	if d == nil {
		return
	}
	d.defaultPath = aPath
	d.key = path.Join(aPath, DefaultSecretKeyFile)
	d.secrets.Envs = make(map[string]*yamlSecure)
	for _, curEnv := range []string{Global, env} {
		d.SetDefaultFile(curEnv)
	}
	d.curEnv = env
}

// DefineDefaultCredFileName define the internal credential path file for a specific environment.
func (d *Secure) DefineDefaultCredFileName(aPath, env string) string {
	if d == nil {
		return ""
	}
	if env == Global {
		return path.Join(aPath, DefaultCredsFile)
	}
	return path.Join(aPath, env+"-"+DefaultCredsFile)
}

// DefineDefaultSecretFileName define the internal credential path file for a specific environment.
func (d *Secure) DefineDefaultSecretFileName(aPath, env string) string {
	if d == nil {
		return ""
	}
	if env == Global {
		return path.Join(aPath, DefaultSecretFile)
	}
	return path.Join(aPath, env+"-"+DefaultSecretFile)
}

// SetDefaultFile
func (d *Secure) SetDefaultFile(env string) {
	if d == nil || d.secrets.Envs == nil {
		return
	}
	data := yamlSecure{
		Version:   CredsVersion,
		file:      path.Clean(d.DefineDefaultCredFileName(d.defaultPath, env)),
		credFile:  path.Clean(d.DefineDefaultSecretFileName(d.defaultPath, env)),
		file_path: d.defaultPath,
		s:         &d.secrets,
	}
	d.secrets.Envs[env] = &data
	return
}

// SetFile load a single file for the env given.
// if env is 'global', so data is considered as valid for all environment.
func (d *Secure) SetFile(filePath, env string) {
	if d == nil || d.secrets.Envs == nil {
		return
	}
	data := yamlSecure{
		Version:   CredsVersion,
		file:      path.Clean(filePath),
		file_path: path.Dir(filePath),
	}
	d.secrets.Envs[env] = &data
}

// SetForjValue set a value in Forj section.
func (d *Secure) SetForjValue(env, source, key string, value *ForjValue) (_ bool, _ error) {
	if d == nil {
		return
	}
	if v, found := d.secrets.Envs[env]; found {
		if v.SetForjValue(source, key, value) {
			d.secrets.Envs[env] = v
			d.updated = true
		}
		return d.updated, nil
	}
	return d.updated, fmt.Errorf("Credential env '%s' nt found", env)
}

// SetForjValue set a value in Forj section.
func (d *Secure) GetForjValue(env, key string) (_ string, _ bool) {
	if d == nil {
		return
	}
	if v, found := d.secrets.Envs[env]; found {
		value, foundValue := v.GetForjValue(key)
		return value.value, foundValue
	}
	return
}

// SetObjectValue set object value
func (d *Secure) SetObjectValue(env, source, obj_name, instance_name, key_name string, value *ObjectsValue) (_ bool) {
	if d == nil {
		return
	}
	if v, found := d.secrets.Envs[env]; found {
		if v.setObjectValue(source, obj_name, instance_name, key_name, value) {
			d.updated = true
			d.secrets.Envs[env] = v
			return true
		}
	}
	return
}

// UnsetObjectValue remove the object value
func (d *Secure) UnsetObjectValue(env, source, objName, instanceName, keyName string) (_ bool) {
	if d == nil {
		return
	}
	if v, found := d.secrets.Envs[env]; found {
		if v.unsetObjectValue(source, objName, instanceName, keyName) {
			d.updated = true
			d.secrets.Envs[env] = v
			return true
		}
	}
	return
}

// GetGlobalString return a string representation of the value.
func (d *Secure) GetGlobalString(objName, instanceName, keyName string) (value string, found bool, source, env string) {
	if d == nil {
		return
	}
	for _, env = range []string{Global} {
		if v, isFound := d.secrets.Envs[env]; isFound {
			if value, found, source = v.getString(objName, instanceName, keyName); found {
				return
			}
		}
	}
	return "", false, "", ""
}

// GetString return a string representation of the value whatever resource is providing the data (forjj data, files, ...)
func (d *Secure) GetString(objName, instanceName, keyName string) (value string, found bool, source, env string) {
	if d == nil {
		return
	}
	for _, env = range []string{d.curEnv, Global} {
		if v, isFound := d.secrets.Envs[env]; isFound {
			if value, found, source = v.getString(objName, instanceName, keyName); found {
				return
			}
		}
	}
	return "", false, "", ""
}

// Get value of the object instance key...
func (d *Secure) Get(objName, instanceName, keyName string) (value *ObjectsValue, found bool, source, env string) {
	if d == nil {
		return
	}
	for _, env = range []string{d.curEnv, Global} {
		if v, isFound := d.secrets.Envs[env]; isFound {
			if value, found, source = v.get(objName, instanceName, keyName); found {
				return
			}
		}
	}
	return nil, false, "", ""
}

// GetObjectInstance return the instance data
func (d *Secure) GetObjectInstance(objName, instanceName string) (values map[string]*ObjectsValue) {
	if d == nil {
		return
	}
	if v, found := d.secrets.Envs[Global]; found {
		values = v.getObjectInstance(objName, instanceName)
		if v, found = d.secrets.Envs[d.curEnv]; found {
			if values == nil {
				return v.getObjectInstance(objName, instanceName)
			}
			for name, value := range v.getObjectInstance(objName, instanceName) {
				values[name] = value
			}
		}
	}
	return
}

func (d *Secure) GetSecrets(env string) (result *Secrets) {
	if d == nil {
		return
	}
	result = NewSecrets()
	result.Envs[Global] = d.secrets.Envs[Global]
	result.Envs[env] = d.secrets.Envs[env]
	return
}
