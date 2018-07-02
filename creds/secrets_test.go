package creds

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_NewSecrets(t *testing.T) {
	t.Log("Expecting NewSecrets to properly intialized the Secrets object.")

	// ------------- call the function
	v := NewSecrets()
	// -------------- testing
	if v == nil {
		t.Error("Expected NewSecret() to return a Secret object. Got nil")
	} else if v.Envs == nil {
		t.Error("Expected secret.Envs to be initialized. got nil")
	}
}

func Test_GenerateKey(t *testing.T) {
	t.Log("Expecting GenerateKey to create a random key.")

	var s1 *Secrets

	// ------------- call the function
	ret := s1.GenerateKey()

	// -------------- testing
	if ret == nil {
		t.Error("Expected GenerateKey() to fail. Got nil")
	} else if !strings.Contains(fmt.Sprintf("%s", ret), "Secret object is nil") {
		t.Errorf("Expected GenerateKey() to fail. with proper error. Got %s", ret)
	}

	s1 = NewSecrets()
	// ------------- call the function
	ret = s1.GenerateKey()
	// -------------- testing
	if ret != nil {
		t.Errorf("Expected GenerateKey() to return a Secret object. Got %s", ret)
	} else if s1.key == nil {
		t.Error("Expected GenerateKey() to generate a key. Got nil")
	} else if v := len(s1.key); v != KeySize {
		t.Errorf("Expected GenerateKey() to generate a key sized at %d. Got %d", KeySize, v)
	} else if s1.key64 == "" {
		t.Errorf("Expected GenerateKey() to generate a key in base64. Empty string.")
	} else if v1, err := base64.StdEncoding.DecodeString(s1.key64); err != nil {
		t.Errorf("Expected GenerateKey() to generate a valid key in base64. failure with error '%s'", err)
	} else if !reflect.DeepEqual(v1, s1.key) {
		t.Error("Expected GenerateKey() to generate a valid key in base64. decoded data is different'")
	}

	analyze := make([][]byte, 10)
	for index := 0; index < 10; index++ {
		analyze[index] = s1.key
		s1.GenerateKey()
	}

	iCount := 0
	for index := 0; index < 10; index++ {
		for index2 := 0; index2 < 10; index2++ {
			if index == index2 {
				continue
			}
			if reflect.DeepEqual(analyze[index], analyze[index2]) {
				iCount++
			}
		}
		analyze[index] = s1.key
		s1.GenerateKey()
	}
	if iCount > 2 {
		t.Error("Expected GenerateKey() to generate a random key. At least found 2 identical keys.")
	}
}

func Test_SaveKey(t *testing.T) {
	t.Log("Expecting SaveKey to properly save the secret key.")

	s := NewSecrets()
	if s == nil {
		t.Error("Test context failure. secret is nil")
		return
	}
	err := s.GenerateKey()
	if err != nil {
		t.Errorf("Test context failure. key generate error: %s", err)
		return
	}

	var fd *os.File
	fd, err = ioutil.TempFile("", "readKey-")

	fileName := fd.Name()
	fd.Close()
	os.Remove(fileName)
	// ------------- call the function
	err = s.SaveKey(fileName)
	defer os.Remove(fileName)
	// -------------- testing
	if err != nil {
		t.Errorf("Expected SaveKey() to have no error. Got '%s'", err)
	} else if info, err1 := os.Stat(fileName); err != nil {
		t.Errorf("Expected SaveKey() to return nil. Got %s", err1)
	} else if int(info.Size()) != len(s.key64) {

	}
}

func Test_LoadKey(t *testing.T) {
	t.Log("Expecting LoadKey to properly load a saved Secret key.")

	s := NewSecrets()
	if s == nil {
		t.Error("Test context failure. secret is nil")
		return
	}
	err := s.GenerateKey()
	if err != nil {
		t.Errorf("Test context failure. key generate error: %s", err)
		return
	}

	var fd *os.File
	fd, err = ioutil.TempFile("", "readKey-")

	fileName := fd.Name()
	fd.Close()
	os.Remove(fileName)
	err = s.SaveKey(fileName)
	defer os.Remove(fileName)

	sl := NewSecrets()
	if sl == nil {
		t.Error("Test context failure. secret is nil")
		return
	}

	// ------------- call the function
	err = sl.ReadKey(fileName)

	// -------------- testing
	if err != nil {
		t.Errorf("Expected LoadKey() to have no error. Got '%s'", err)
	} else if sl.key64 == "" {
		t.Error("Expected LoadKey() to load the key saved in base64. Got empty string")
	} else if sl.key64 != s.key64 {
		t.Error("Expected LoadKey() to load the key saved in base64. Got another key")
	} else if sl.key == nil {
		t.Error("Expected LoadKey() to load the key saved. Got nil")
	} else if l := len(sl.key); l != KeySize {
		t.Errorf("Expected LoadKey() to load the key saved. But key size is incorrect. %d != %d", KeySize, l)
	} else if !reflect.DeepEqual(sl.key, s.key) {
		t.Errorf("Expected LoadKey() to load the key saved. But key is not the saved one. %d != %d", KeySize, l)
	}
}

func Test_Key64(t *testing.T) {
	t.Log("Expecting Key64 to properly return the key in base64.")

	var s *Secrets

	// ------------- call the function
	val := s.Key64()

	// -------------- testing
	if val != "" {
		t.Errorf("Expected Key64() to return an empty string. Got '%s'", val)
	}

	// -------------- Update context
	s = NewSecrets()
	if s == nil {
		t.Error("Test context failure. secret is nil")
		return
	}

	// ------------- call the function
	val = s.Key64()

	// -------------- testing
	if val != "" {
		t.Errorf("Expected Key64() to return an empty string. Got '%s'", val)
	}

	// -------------- Update context
	err := s.GenerateKey()
	if err != nil {
		t.Errorf("Test context failure. key generate error: %s", err)
		return
	}

	// ------------- call the function
	val = s.Key64()

	// -------------- testing
	if val == "" {
		t.Error("Expected Key64() to return an non empty string. Is empty")
	} else if val != s.key64 {
		t.Error("Expected Key64() to return the internal base64 key. Got a different one")
	}

}

func Test_Export(t *testing.T) {
	t.Log("Expecting Export to properly return the secret encrypted.")

	var s *Secrets

	// ------------- call the function
	// secret is not defined.
	ret, err := s.Export()

	// -------------- testing
	if ret != nil {
		t.Error("Expected Export() to fail. Got some data")
	} else if !strings.Contains(fmt.Sprintf("%s", err), "Secret object is nil") {
		t.Errorf("Expected Export() to fail. with proper error. Got %s", err)
	}

	// -------------- Update context
	s = NewSecrets()
	if s == nil {
		t.Error("Test context failure. secret is nil")
		return
	}

	// ------------- call the function
	// Empty secret
	ret, err = s.Export()

	// -------------- testing
	if ret != nil {
		t.Error("Expected Export() to fail. Got some data")
	} else if !strings.Contains(fmt.Sprintf("%s", err), "Key is missing") {
		t.Errorf("Expected Export() to fail. with proper error. Got '%s'", err)
	}

	// -------------- Update context
	err = s.GenerateKey()
	if err != nil {
		t.Errorf("Test context failure. key generate error: %s", err)
		return
	}

	// ------------- call the function
	ret, err = s.Export()

	// -------------- testing
	if err != nil {
		t.Errorf("Expected Export() to succeed but it fail with '%s'", err)
	} else if ret == nil {
		t.Error("Expected Export() to succeed but it fail with no data")
	}
}

func Test_SetKey64(t *testing.T) {
	t.Log("Expecting SetKey64 to properly set the secret key.")

	s := NewSecrets()
	if s == nil {
		t.Error("Test context failure. secret is nil")
		return
	}
	err := s.GenerateKey()
	if err != nil {
		t.Errorf("Test context failure. key generate error: %s", err)
		return
	}

	sk := NewSecrets()
	if sk == nil {
		t.Error("Test context failure. secret sk is nil")
		return
	}

	// ------------- call the function
	err = sk.SetKey64(s.Key64())

	// -------------- testing
	if err != nil {
		t.Errorf("Expected SetKey64() to have no error. Got '%s'", err)
	} else if sk.key64 == "" {
		t.Error("Expected SetKey64() to set the key saved in base64. Got empty string")
	} else if sk.key64 != s.key64 {
		t.Error("Expected SetKey64() to set the key saved in base64. Got another key")
	} else if sk.key == nil {
		t.Error("Expected SetKey64() to set the key given. Got nil")
	} else if l := len(sk.key); l != KeySize {
		t.Errorf("Expected SetKey64() to set the key given. But key size is incorrect. %d != %d", KeySize, l)
	} else if !reflect.DeepEqual(sk.key, s.key) {
		t.Errorf("Expected SetKey64() to set the key given. But key is not the saved one. %d != %d", KeySize, l)
	}
}

func Test_Import(t *testing.T) {
	t.Log("Expecting Import to properly import the secret encrypted.")

	s := NewSecrets()
	if s == nil {
		t.Error("Test context failure. secret is nil")
		return
	}
	err := s.GenerateKey()
	if err != nil {
		t.Errorf("Test context failure. key generate error: %s", err)
		return
	}

	s.Envs["test"] = yamlSecure{
		Version: "V1",
	}

	var dataEnc []byte
	dataEnc, err = s.Export()
	if err != nil {
		t.Error("Test context failure. secret is nil")
		return
	}

	si := NewSecrets()
	if si == nil {
		t.Error("Test context failure. secret is nil")
		return
	}

	si.SetKey64(s.Key64())
	// ------------- call the function
	err = si.Import(dataEnc)

	// -------------- testing
	if err != nil {
		t.Errorf("Expected Import() to have no error. Got '%s'", err)
	} else if v, found := si.Envs["test"]; !found {
		t.Error("Expected Import() to have structure imported properly. Env['test'] not found.")
	} else if v.Version != "V1" {
		t.Errorf("Expected Import() to have structure imported properly. Env['test'].Version is not '%s'. got '%s'", "V1", v.Version)
	}
}
