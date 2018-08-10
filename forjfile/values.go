package forjfile

// TODO: Add a function/cap to forjj to generate Forjfile with Default values (ForjValueSelectDefault = true)
var ForjValueSelectDefault bool

type ForjValues map[string]ForjValue

// Map returns a map[string]string of all values stored.
//
// It returns values setup with Set and if not found, 
// returns Default value.
func (v ForjValues) Map() (values map[string]string) {
	values = make(map[string]string)
	for key, value := range v {
		if v1, f1 := value.get_selected() ; f1 {
			values[key] = v1
		}
	}
	return
}

func (v ForjValues) MarshalYAML() (interface{}, error) {
	values := make(map[string]string)

	for key, value := range v {
		if v1, f1 := value.get_selected() ; f1 {
			values[key] = v1
		}
	}
	return values, nil
}
