package creds

import (
	"fmt"

	"github.com/forj-oss/goforjj"
)

// ObjectValue describe the Objects keys value
type Value struct {
	value *goforjj.ValueStruct // value.
	// If source == `forjj` => real value
	// If source == `file` => Path the a file containing the value
	// Else => address of the data, with eventually a collection of resources to help getting the data from the address given.

	resource map[string]string // Collection of resources to identify where is the data and how to access it
	source   string            // Source of the data. Can be `forjj`, `file` or an external service like `plugin:vault`
	s        *Secrets          // Ref to the main struct address
}

const (
	// Internal represents the forjj internal storage
	Internal = "internal"
	// Link represents the file link storage.
	Link = "link"
)

// NewValue creates a new Value object, initialized with a ValueStruct if needed.
func NewValue(source string, value *goforjj.ValueStruct) (ret *Value) {
	ret = new(Value)
	ret.Set(source, value)
	return
}

func (v *Value) init() {
	if v == nil {
		return
	}
	if v.resource == nil {
		v.resource = make(map[string]string)
	}
}

// clone creates a clone of the Value object and attach the Secrets object ref
func (v Value) clone(s *Secrets) *Value {
	v.setSecrets(s)
	v.value = goforjj.NewValueStruct(v.value)
	return &v
}

// Clone create a YamlValue struct duplicating Value.
func (v Value) Clone() (ret *YamlValue) {
	ret = new(YamlValue)
	ret.Value = goforjj.NewValueStruct(v.value)
	ret.Resource = v.resource
	ret.Source = v.source
	return
}

// copy a Value content to an existing Value object. The Secrets object ref is unchanged.
func (v *Value) copyFrom(copy *Value) (updated bool) {
	if v == nil {
		return
	}
	if v.value == nil {
		v.Set(copy.source, goforjj.NewValueStruct(copy.value))
		updated = true
		v.resource = copy.resource
	} else {
		updated = !copy.value.Equal(v.value)
		// Set may update the list of resources depending on the type and handler attached.
		v.Set(copy.source, copy.value)
	}
	return
}

// setSecret stores the main struct address.
func (v *Value) setSecrets(s *Secrets) {
	if v == nil {
		return
	}
	v.s = s
}

// Set source andvalue of a Value instance
func (v *Value) Set(source string, value *goforjj.ValueStruct) (err error) {
	if v == nil {
		return
	}

	v.init()

	if source == "" {
		source = Internal
	}
	if source == Internal || v.s == nil {
		v.SetValue(value)
		v.source = source
		return
	}
	if v.s != nil {
		if setter, found := v.s.setter[source]; !found {
			err = fmt.Errorf("'%s' is an unknown getter type", v.source)
			return
		} else {
			v.source = source
			v.resetResources()
			return setter(v, value)
		}
	}
	return
}

// SetValue set the value of a Value instance
func (v *Value) SetValue(value interface{}) {
	if v == nil {
		return
	}

	v.init()

	if v.value == nil {
		v.value = goforjj.NewValueStruct(nil)
		v.source = Internal
	}
	v.value.Set(value)
	if v.resource == nil {
		v.resource = make(map[string]string)
	}
}

// SetSource set the source information of a Value instance
func (v *Value) SetSource(source string) {
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

// GetSource get the source information of a Value instance
func (v *Value) GetSource() string {
	if v == nil {
		return ""
	}
	if v.value == nil {
		return ""
	}
	if v.source == "" {
		return Internal
	}

	return v.source
}

// GetString source and value of a ForjValue instance
func (v *Value) GetString() (_ string, err error) {
	if v == nil {
		return
	}

	if v.source == Internal || v.source == "" {
		return v.value.GetString(), nil
	}

	if v.s != nil {
		if getter, found := v.s.getter[v.source]; !found {
			err = fmt.Errorf("'%s' is an unknown getter type", v.source)
			return
		} else {
			return getter(v.Clone())
		}
	}
	return v.value.GetString(), nil
}

// ------------- Resource management

func (v *Value) resetResources() {
	if v == nil {
		return
	}

	v.resource = make(map[string]string)
}

// GetResource return the resource value
func (v *Value) GetResource(key string) (value string, found bool) {
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
func (v *Value) AddResource(key, value string) {
	if v == nil {
		return
	}
	if v.resource == nil {
		v.resource = make(map[string]string)
	}

	v.resource[key] = value
}

// ----------------- Value versioning behavior.

// MarshalYAML encode the object in ValueStruct output
func (v Value) MarshalYAML() (interface{}, error) {
	// Version 0.2
	value := YamlValue{
		Value:    v.value,
		Resource: v.resource,
		Source:   v.source,
	}
	return value, nil
}

// UnmarshalYAML decode the flow as a ValueStruct
func (v *Value) UnmarshalYAML(unmarchal func(interface{}) error) (err error) {
	if v.value == nil {
		v.value = new(goforjj.ValueStruct)
	}

	if data.Version == "0.1" {
		return v.value.UnmarshalYAML(unmarchal)
	}

	// Version 0.2
	data := new(YamlValue)
	err = unmarchal(data)
	v.value = data.Value
	v.resource = data.Resource
	v.source = data.Source
	return
}
