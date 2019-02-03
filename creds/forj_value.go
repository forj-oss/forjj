package creds

// ForjjValue describe the Forjj keys value
type ForjValue struct {
	value string // value.
	// If source == `forjj` => real value
	// If source == `file` => Path the a file containing the value
	// Else => address of the data, with eventually a collection of resources to help getting the data from the address given.
	resource map[string]string // Collection of resources to identify where is the data and how to access it
	source   string            // Source of the data. Can be `forjj`, `file` or an external service like `plugin:vault`
}

type yamlForjValue struct {
	Value    string
	Resource map[string]string
	Source   string
}

func NewForjValue(source, value string) (ret *ForjValue) {
	ret = new(ForjValue)
	ret.Set(source, value)
	return
}

// Set source andvalue of a ForjValue instance
func (v *ForjValue) Set(source, value string) {
	if v == nil {
		return
	}
	v.value = value
	v.source = source
	if v.resource == nil {
		v.resource = make(map[string]string)
	}
}

// AddResource adds resources information to the data given
func (v *ForjValue) AddResource(key, value string) {
	if v == nil {
		return
	}
	if v.resource == nil {
		v.resource = make(map[string]string)
	}

	v.resource[key] = value
}

// MarshalYAML encode the object in ValueStruct output
func (v ForjValue) MarshalYAML() (interface{}, error) {
	// Version 0.2
	value := yamlForjValue{
		Value:    v.value,
		Resource: v.resource,
		Source:   v.source,
	}

	return value, nil
}

// UnmarshalYAML decode the flow as a ValueStruct
func (v *ForjValue) UnmarshalYAML(unmarchal func(interface{}) error) (err error) {
	if data.Version == "0.1" {
		return unmarchal(&v.value)
	}

	// Version 0.2
	data := new(yamlForjValue)
	err = unmarchal(data)
	v.resource = data.Resource
	v.source = data.Source
	v.value = data.Value
	return

}
