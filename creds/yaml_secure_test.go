package creds

import (
	"strings"
	"testing"

	"github.com/forj-oss/goforjj"
)

func Test_YamlSecure_SetForjValue(t *testing.T) {

	t.Log("Expecting yamlSecure.SetForjValue to set values in Forj section.")

	s := yamlSecure{}
	const (
		key1   = "key1"
		value1 = "value1"
	)

	// ------------- call the function
	s.SetForjValue(key1, value1)

	// -------------- testing
	if s.Forj == nil {
		t.Error("Expected s.Forj to be set. Got nil")
	} else if l := len(s.Forj); l != 1 {
		t.Errorf("Expected s.Forj to have 1 element. Got %d.", l)
	} else if v, found := s.Forj[key1]; !found {
		t.Errorf("Expected s.Forj[%s] to exist. Not found", key1)
	} else if v != value1 {
		t.Errorf("Expected s.Forj[%s] to be '%s'. Got '%s'", key1, value1, v)
	}
}

func Test_YamlSecure_setObjectValue(t *testing.T) {

	t.Log("Expecting yamlSecure.setObjectValue to set values in Objects section.")

	s := yamlSecure{}
	const (
		object1   = "object1"
		object2   = "object2"
		instance1 = "instance1"
		instance2 = "instance2"
		key1      = "key1"
		value1    = "value1"
		value2    = "value2"
	)

	// ------------- call the function
	value := new(goforjj.ValueStruct)
	value.Set(value1)
	result := s.setObjectValue(object1, instance1, key1, value)

	// -------------- testing
	if s.Objects == nil {
		t.Error("Expected s.Objects to be set. Got nil")
	} else if l1 := len(s.Objects); l1 != 1 {
		t.Errorf("Expected s.Objects to have 1 element. Got %d.", l1)
	} else if v1, found := s.Objects[object1]; !found {
		t.Errorf("Expected s.Objects[%s] to exist. Not found", object1)
	} else if v1 == nil {
		t.Errorf("Expected s.Objects[%s] to be set. Got nil", object1)
	} else if l2 := len(v1); l2 != 1 {
		t.Errorf("Expected s.Objects to have 1 element. Got %d.", l2)
	} else if v2, found := v1[instance1]; !found {
		t.Errorf("Expected s.Objects[%s][%s] to exist. Not found", object1, instance1)
	} else if v2 == nil {
		t.Errorf("Expected s.Objects[%s][%s] to be set. Got nil", object1, instance1)
	} else if l3 := len(v2); l3 != 1 {
		t.Errorf("Expected s.Objects to have 1 element. Got %d.", l3)
	} else if v3, found := v2[key1]; !found {
		t.Errorf("Expected s.Objects[%s][%s][%s] to exist. Not found", object1, instance1, key1)
	} else if v3 == nil {
		t.Errorf("Expected s.Objects[%s][%s][%s] to be set. Got nil", object1, instance1, key1)
	} else if v4 := v3.GetString(); v4 != value1 {
		t.Errorf("Expected s.Objects[%s][%s][%s] to be '%s'. Got '%s'", object1, instance1, key1, v4, value1)
	} else if !result {
		t.Error("Expected setObjectValue to return true. got false.")
	}

	// ------------- call the function
	result = s.setObjectValue(object1, instance1, key1, value)

	// -------------- testing
	if result {
		t.Error("Expected setObjectValue to return false. got true.")
	}

	// ------------- call the function
	value.Set(value2)
	result = s.setObjectValue(object1, instance1, key1, value)

	// -------------- testing
	if !result {
		t.Error("Expected setObjectValue to return true. got false.")
	}

	// ------------- call the function
	result = s.setObjectValue(object2, instance1, key1, value)

	// -------------- testing
	if s.Objects == nil {
		t.Error("Expected s.Objects to be set. Got nil")
	} else if l1 := len(s.Objects); l1 != 2 {
		t.Errorf("Expected s.Objects to have 1 element. Got %d.", l1)
	} else if _, found := s.Objects[object2]; !found {
		t.Errorf("Expected s.Objects[%s] to exist. Not found", object1)
	}

	// Change value context
	value.Set(value2)
	// ------------- call the function
	// Set a new instance, key and value
	result = s.setObjectValue(object2, instance2, key1, value)

	// -------------- testing
	if s.Objects == nil {
		t.Error("Expected s.Objects to be set. Got nil")
	} else if l1 := len(s.Objects); l1 != 2 {
		t.Errorf("Expected s.Objects to have 1 element. Got %d.", l1)
	} else if v1, found := s.Objects[object2]; !found {
		t.Errorf("Expected s.Objects[%s] to exist. Not found", object1)
	} else if v2, found := v1[instance2]; !found {
		t.Errorf("Expected s.Objects[%s][%s] to exist. Not found", object1, instance1)
	} else if v2 == nil {
		t.Errorf("Expected s.Objects[%s][%s] to be set. Got nil", object1, instance1)
	} else if l3 := len(v2); l3 != 1 {
		t.Errorf("Expected s.Objects to have 1 element. Got %d.", l3)
	} else if v3, found := v2[key1]; !found {
		t.Errorf("Expected s.Objects[%s][%s][%s] to exist. Not found", object1, instance1, key1)
	} else if v3 == nil {
		t.Errorf("Expected s.Objects[%s][%s][%s] to be set. Got nil", object1, instance1, key1)
	} else if v4 := v3.GetString(); v4 != value2 {
		t.Errorf("Expected s.Objects[%s][%s][%s] to be '%s'. Got '%s'", object1, instance1, key1, v4, value2)
	} else if !result {
		t.Error("Expected setObjectValue to return true. got false.")
	}
}

func Test_YamlSecure_get(t *testing.T) {

	t.Log("Expecting yamlSecure.get to set values in Objects section.")

	s := yamlSecure{}
	const (
		object1   = "object1"
		object2   = "object2"
		object3   = "object3"
		instance1 = "instance1"
		key1      = "key1"
		key2      = "key2"
		value1    = "value1"
		value2    = "value2"
	)

	value := new(goforjj.ValueStruct)
	value.Set(value1)
	s.setObjectValue(object1, instance1, key1, value)
	// ------------- call the function

	result, found := s.get(object1, instance1, key1)
	// -------------- testing
	if result == nil {
		t.Error("Expected result to be set. Got nil")
	} else if !found {
		t.Error("Expected to have found to be true. got false")
	} else if !result.Equal(value) {
		t.Error("Expected result to be equal to original value. got false")
	}

	// ------------- call the function
	result, found = s.get(object2, instance1, key1)

	// -------------- testing
	if result != nil {
		t.Error("Expected result to be nil. Got a result")
	} else if found {
		t.Error("Expected to have found to be false. got true")
	}

	// ------------- Update context
	value.Set(value2)

	// ------------- call the function
	result, found = s.get(object1, instance1, key1)

	// -------------- testing
	if result.Equal(value) {
		t.Error("Expected result to NOT be equal to original value. got true.")
	}

	// ------------- Update context
	value = result
	result.Set(value2)

	// ------------- call the function
	result, found = s.get(object1, instance1, key1)

	// -------------- testing
	if result.Equal(value) {
		t.Error("Expected result to NOT be equal to original value. got true.")
	}

	yamlData := `---
`

	const (
		appObj = "app"
		appIns = "jenkins"
		appKey = "aws-iam-arn-slave"
	)

	yamlData = `---
objects:
`
	// update context
	r := strings.NewReader(yamlData)
	err := s.iLoad(r)

	if err != nil {
		t.Errorf("Context error: %s", err)
		return
	}
	// ------------- call the function
	result, found = s.get(appObj, appIns, appKey)

	if found {
		t.Errorf("Expect to not found any values. Got one.")
	} else if result != nil {
		t.Errorf("Expect to not return any values. Got one.")
	}

	yamlData = `---
objects:
  app:
`
	// update context
	r = strings.NewReader(yamlData)
	err = s.iLoad(r)

	if err != nil {
		t.Errorf("Context error: %s", err)
		return
	}
	// ------------- call the function
	result, found = s.get(appObj, appIns, appKey)

	if found {
		t.Errorf("Expect to not found any values. Got one.")
	} else if result != nil {
		t.Errorf("Expect to not return any values. Got one.")
	}

	yamlData = `---
objects:
  app:
    jenkins:
`
	// update context
	r = strings.NewReader(yamlData)
	err = s.iLoad(r)

	if err != nil {
		t.Errorf("Context error: %s", err)
		return
	}
	// ------------- call the function
	result, found = s.get(appObj, appIns, appKey)

	if found {
		t.Errorf("Expect to not found any values. Got one.")
	} else if result != nil {
		t.Errorf("Expect to not return any values. Got one.")
	}

	yamlData = `---
objects:
  app:
    jenkins:
      aws-iam-arn-slave:
`
	// update context
	r = strings.NewReader(yamlData)
	err = s.iLoad(r)

	if err != nil {
		t.Errorf("Context error: %s", err)
		return
	}
	// ------------- call the function
	result, found = s.get(appObj, appIns, appKey)

	if found {
		t.Errorf("Expect to not found any values. Got one.")
	} else if result != nil {
		t.Errorf("Expect to not return any values. Got one.")
	}

}
