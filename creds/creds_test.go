package creds

import (
	"path"
	"testing"
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
	t.Log("Expecting Save to save files. NOT TESTED.")
}
