package creds

import (
	"reflect"
	"testing"

	"github.com/forj-oss/goforjj"
	"github.com/stretchr/testify/assert"
)

func Test_NewObjectsValue(t *testing.T) {
	t.Log("Expecting NewForjValue to properly initialized the ForjValue object.")
	assert := assert.New(t)
	value := goforjj.NewValueStruct("value")

	// ------------- call the function
	v := NewObjectsValue("forjj", value)

	// -------------- testing
	when := "when a new ObjectsValue is set"
	if assert.NotNilf(v, "Expected ObjectsValue to be returned %s", when) {
		if assert.NotNilf(v.value, "Expected ObjectValue to have a valid ValueStruct %s", when) {
			assert.Truef(v.value.Equal(value), "Expected value to be properly set %s", when)
			assert.Equalf("forjj", v.source, "Expected source to be set properly %s", when)
			if assert.NotNilf(v.resource, "Expected resource to be initialized %s", when) {
				assert.Lenf(v.resource, 0, "Expected resource to be empty %s", when)
			}
		}
	}
}

func Test_ObjectsValue_Set(t *testing.T) {
	t.Log("Expecting ForjValue.Set to properly set the ForjValue object.")
	assert := assert.New(t)

	v := ObjectsValue{}
	value := goforjj.NewValueStruct("value")

	// ------------- call the function
	v.Set("forjj", value)

	// -------------- testing
	when := "when an empty ForjValue is set"
	if assert.NotNilf(v.value, "Expected ObjectValue to have a valid ValueStruct %s", when) {
		assert.Truef(v.value.Equal(value), "Expected value to be properly set %s", when)
		assert.Equalf("forjj", v.source, "Expected source to be set properly %s", when)
		if assert.NotNilf(v.resource, "Expected resource to be initialized %s", when) {
			assert.Lenf(v.resource, 0, "Expected resource to be empty %s", when)
		}
	}

	// -------------- Update context
	value.Set("value2")

	// -------------- testing
	when = "when the ValueStruct value has been updated outside"
	if assert.NotNilf(v.value, "Expected ObjectValue to have a valid ValueStruct %s", when) {
		assert.Falsef(v.value.Equal(value), "Expected value to be updated %s", when)
		assert.Equalf("forjj", v.source, "Expected source to be set properly %s", when)
		if assert.NotNilf(v.resource, "Expected resource to be initialized %s", when) {
			assert.Lenf(v.resource, 0, "Expected resource to be empty %s", when)
		}
	}

	// ------------- call the function
	v.Set("blabla", value)

	// -------------- testing
	when = "when an existing ObjectsValue is set"
	if assert.NotNilf(v, "Expected ObjectsValue to be returned %s", when) {
		assert.Truef(v.value.Equal(value), "Expected value to be properly set to new value %s", when)
		assert.Equalf("blabla", v.source, "Expected source to be set properly to new value %s", when)
		if assert.NotNilf(v.resource, "Expected resource to be initialized %s", when) {
			assert.Lenf(v.resource, 0, "Expected resource to be empty %s", when)
		}
	}

	// -------------- Update context
	v1 := NewObjectsValue("forjj", goforjj.NewValueStruct("value"))

	// ------------- call the function
	v1.Set("blabla", value)

	// -------------- testing
	when = "when an existing ObjectsValue is set"
	if assert.NotNilf(v1, "Expected ObjectsValue to be returned %s", when) {
		assert.Truef(v.value.Equal(value), "Expected value to be properly set to new value %s", when)
		assert.Equalf("blabla", v1.source, "Expected source to be set properly to new value %s", when)
		if assert.NotNilf(v1.resource, "Expected resource to be initialized %s", when) {
			assert.Lenf(v1.resource, 0, "Expected resource to be empty %s", when)
		}
	}
}

func Test_ObjectsValue_AddResource(t *testing.T) {
	t.Log("Expecting ForjValue.AddResource to properly set the ForjValue object resource list.")
	test := assert.New(t)

	v := ObjectsValue{}
	r := map[string]string{
		"key": "value",
	}

	// ------------- call the function
	v.AddResource("key", "value")

	// -------------- testing
	when := "when a resource is added on an empty ForjValue."
	test.Emptyf(v.value, "Expected value to stay empty %s", when)
	test.Emptyf(v.source, "Expected source to stay empty %s", when)
	if test.NotNilf(v.resource, "Expected resource to be initialized %s", when) {
		test.Lenf(v.resource, 1, "Expected resource to be empty %s", when)
		test.Truef(reflect.DeepEqual(v.resource, r), "Expected resource to be conform %s", when)
	}
	// -------------- Update context
	r["key"] = "value2"

	// ------------- call the function
	v.AddResource("key", "value2")

	// -------------- testing
	when = "when a resource is updated."
	test.Emptyf(v.value, "Expected value to stay empty %s", when)
	test.Emptyf(v.source, "Expected source to stay empty %s", when)
	if test.NotNilf(v.resource, "Expected resource to be initialized %s", when) {
		test.Lenf(v.resource, 1, "Expected resource to be empty %s", when)
		test.Truef(reflect.DeepEqual(v.resource, r), "Expected resource to be conform %s", when)
	}
}
