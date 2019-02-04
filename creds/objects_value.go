package creds

import "github.com/forj-oss/goforjj"

// ObjectValue describe the Objects keys value
type ObjectsValue struct {
	value *goforjj.ValueStruct // value.
	// If source == `forjj` => real value
	// If source == `file` => Path the a file containing the value
	// Else => address of the data, with eventually a collection of resources to help getting the data from the address given.

	resource map[string]string // Collection of resources to identify where is the data and how to access it
	source   string            // Source of the data. Can be `forjj`, `file` or an external service like `plugin:vault`
}

const (
	internal = "internal"
)

// NewObjectsValue creates a new ObjectValue object, initialized with a ValueStruct if needed.
func NewObjectsValue(source string, value *goforjj.ValueStruct) (ret *ObjectsValue) {
	ret = new(ObjectsValue)
	ret.Set(source, value)
	return
}

func (v *ObjectsValue) init() {
	if v == nil {
		return
	}
	if v.resource == nil {
		v.resource = make(map[string]string)
	}
}

// Set source andvalue of a ObjectsValue instance
func (v *ObjectsValue) Set(source string, value *goforjj.ValueStruct) {
	if v == nil {
		return
	}

	v.init()

	v.SetValue(value)
	v.source = source
}

// SetValue set the value of a ObjectsValue instance
func (v *ObjectsValue) SetValue(value interface{}) {
	if v == nil {
		return
	}

	v.init()

	if v.value == nil {
		v.value = goforjj.NewValueStruct(nil)
		v.source = internal
	}
	v.value.Set(value)
	if v.resource == nil {
		v.resource = make(map[string]string)
	}
}

// SetSource set the source information of a ObjectsValue instance
func (v *ObjectsValue) SetSource(source string) {
	if v == nil {
		return
	}
	if v.value == nil {
		v.value = goforjj.NewValueStruct(nil)
	}
	v.source = source
	if v.resource == nil {
		v.resource = make(map[string]string)
	}
}

// GetSource get the source information of a ObjectsValue instance
func (v *ObjectsValue) GetSource() string {
	if v == nil {
		return ""
	}
	if v.value == nil {
		return ""
	}
	if v.source == "" {
		return internal
	}

	return v.source
}

// GetString source andvalue of a ForjValue instance
func (v *ObjectsValue) GetString() (_ string) {
	if v == nil {
		return
	}
	return v.value.GetString()
}

// GetResource return the resource value
func (v *ObjectsValue) GetResource(key string) (value string, found bool) {
	if v == nil {
		return
	}
	if v.resource == nil {
		v.resource = make(map[string]string)
	}

	value, found = v.resource[key]
	return
}


// AddResource adds resources information to the data given
func (v *ObjectsValue) AddResource(key, value string) {
	if v == nil {
		return
	}
	if v.resource == nil {
		v.resource = make(map[string]string)
	}

	v.resource[key] = value
}

// MarshalYAML encode the object in ValueStruct output
func (v ObjectsValue) MarshalYAML() (interface{}, error) {
	// Version 0.2
	value := yamlObjectsValue{
		Value:    v.value,
		Resource: v.resource,
		Source:   v.source,
	}
	return value, nil
}

// UnmarshalYAML decode the flow as a ValueStruct
func (v *ObjectsValue) UnmarshalYAML(unmarchal func(interface{}) error) (err error) {
	if v.value == nil {
		v.value = new(goforjj.ValueStruct)
	}

	if data.Version == "0.1" {
		return v.value.UnmarshalYAML(unmarchal)
	}

	// Version 0.2
	data := new(yamlObjectsValue)
	err = unmarchal(data)
	v.value = data.Value
	v.resource = data.Resource
	v.source = data.Source
	return
}
