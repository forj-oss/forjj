package creds

import (
	"path"
	"testing"

	"github.com/forj-oss/goforjj"
)

func TestInitEnvDefaults(t *testing.T) {
	t.Log("Expecting InitEnvDefaults to properly intialized the cred object.")
	d := new(Secure)

	const (
		myPath = "myPath"
		prod   = "prod"
	)

	// Run the function
	t.Log("Running d.InitEnvDefault('%s', '%s')", myPath, prod)
	d.InitEnvDefaults(myPath, prod)

	// Test the result
	if v := d.defaultPath; v != myPath {
		t.Errorf("Expected %s to have '%s'. Got '%s'", "defaultPath", prod, v)
	}
	if d.envs == nil {
		t.Errorf("Expected %s to be set. Got nil", "envs")
	} else if n := len(d.envs); n != 2 {
		t.Errorf("Expected %s to have '%d' elements. Got '%d'", "envs", 2, n)
	} else if _, found := d.envs[Global]; !found {
		t.Errorf("Expected %s to have '%s' elements. Not found.", "envs", Global)
	} else if _, found := d.envs[prod]; !found {
		t.Errorf("Expected %s to have '%s' elements. Not found.", "envs", prod)
	}
}

func TestSetDefaultFile(t *testing.T) {
	t.Log("Expecting SetDefaultFile to properly intialized the cred object.")
	d := new(Secure)

	const (
		myPath = "myPath"
		myFile = "myFile"
		prod   = "prod"
	)

	// Run the function
	d.SetDefaultFile(prod)

	// Test the result
	if d.envs != nil {
		t.Errorf("Expected %s to be unset. Got it set.", "envs")
	}

	// Define the environment
	d.InitEnvDefaults(myPath, prod)

	// Run the function
	d.SetDefaultFile(prod)

	// Test the result
	if d.envs != nil {
		if v, found := d.envs[Global]; found {
			if ref := d.DefineDefaultCredFileName(myPath, Global); v.file != ref {
				t.Errorf("Expected %s to have '%s' elements. Got '%s'.", "envs[global]", ref, v.file)
			}
		}
		if v, found := d.envs[prod]; found {
			if ref := d.DefineDefaultCredFileName(myPath, prod); v.file != ref {
				t.Errorf("Expected %s to have '%s' elements. Got '%s'.", "envs[prod]", ref, v.file)
			}
		}
	}
}

func TestSetFile(t *testing.T) {
	t.Log("Expecting InitSetFile to properly intialized the cred object.")
	d := new(Secure)

	const (
		myPath = "myPath"
		myFile = "myFile"
		prod   = "prod"
	)

	// Run the function
	d.SetFile(path.Join(myPath, myFile), prod)

	// Test the result
	if d.envs != nil {
		t.Errorf("Expected %s to be unset. Got it set.", "envs")
	}

	// Define the environment
	d.InitEnvDefaults(myPath, prod)

	// Run the function
	d.SetFile(path.Join(myPath, myFile), prod)

	// Test the result
	if d.envs != nil {
		if v, found := d.envs[Global]; found {
			if ref := d.DefineDefaultCredFileName(myPath, Global); v.file != ref {
				t.Errorf("Expected %s to have '%s' elements. Got '%s'.", "envs[global]", ref, v.file)
			}
		}
		if v, found := d.envs[prod]; found {
			if ref := path.Join(myPath, myFile); v.file != ref {
				t.Errorf("Expected %s to have '%s' elements. Got '%s'.", "envs[prod]", ref, v.file)
			}
		}
	}
}

func TestDirName(t *testing.T) {
	t.Log("Expecting DirName to return proper dir name. NOT TESTED.")
}

func TestLoad(t *testing.T) {
	t.Log("Expecting DirName to load proper files. NOT TESTED.")
}

func TestSave(t *testing.T) {
	t.Log("Expecting Save to save files. NOT TESTED.")
}

func TestSetForjValue(t *testing.T) {
	t.Log("Expecting SetForjValue to set properly values. ")

	s := Secure{}
	const (
		key1   = "key1"
		value1 = "value1"
		myPath = "myPath"
		myFile = "myFile"
		prod   = "prod"
	)
	s.InitEnvDefaults(myPath, prod)

	// ------------- call the function
	updated, err := s.SetForjValue(prod, key1, value1)

	// -------------- testing
	if !updated {
		t.Error("Expected s.SetForjValue to update it. Got false")
	} else if err != nil {
		t.Errorf("Expected s.SetForjValue to return no error. Got %s.", err)
	} else if v, found := s.GetForjValue(prod, key1); !found {
		t.Errorf("Expected s.Forj[%s] to exist. Not found", key1)
	} else if v != value1 {
		t.Errorf("Expected s.Forj[%s] to be '%s'. Got '%s'", key1, value1, v)
	}

}

func TestSetObjectValue(t *testing.T) {
	t.Log("Expecting SetObjectValue to set properly values.")

	s := Secure{}
	const (
		object1   = "object1"
		object2   = "object2"
		instance1 = "instance1"
		key1      = "key1"
		value1    = "value1"
		value2    = "value2"
		myPath    = "myPath"
		myFile    = "myFile"
		prod      = "prod"
	)
	s.InitEnvDefaults(myPath, prod)

	// ------------- call the function
	value := new(goforjj.ValueStruct)
	value.Set(value1)
	updated := s.SetObjectValue(prod, object1, instance1, key1, value)

	// -------------- testing
	if !updated {
		t.Error("Expected s.SetObjectValue to return updated = true. Got false")
	} else if v, found := s.Get(object1, instance1, key1); !found {
		t.Error("Expected value to be found. Got false")
	} else if v1 := v.GetString(); v1 != value1 {
		t.Errorf("Expected value to be '%s'. Got '%s'", v1, value1)
	}
}
