package forjfile

type ForjValue struct {
	value string
	default_value string
	source string
}

func (v *ForjValue) Set(aValue string) (updated bool) {
	updated = (v.value != aValue)
	v.value = aValue
	return
}

func (v *ForjValue) SetDefault(aDefValue string) (updated bool) {
	updated = (v.default_value != aDefValue)
	v.default_value = aDefValue
	return
}

func (v *ForjValue) Get() string {
	if v.value != "" {
		return v.value
	}
	return v.default_value
}

func (v *ForjValue) Clean(_ string) (_ bool){
	v.value = ""
	v.default_value = ""
	return
}

func (v *ForjValue) IsDefault() (_ bool) {
	if v.default_value != "" && v.value == "" {
		return true
	}
	return
}

func (v *ForjValue)get_selected() (string, bool) {
	if ForjValueSelectDefault && v.value == "" {
		return v.default_value, (v.default_value != "")
	}
	return v.value, (v.value != "")
}

func (v ForjValue) MarshalYAML() (interface{}, error) {
	value := v.value
	return value, nil
}

func (v *ForjValue) UnmarshalYAML(unmarshal func(interface{}) error) (error) {
	if err := unmarshal(&v.value) ; err != nil {
		return err
	}
	return nil
}
