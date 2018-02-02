package forjfile

import "fmt"

type AppsStruct map[string]*AppStruct

func (a AppsStruct)Found(appName string) (* AppStruct, error) {
	if v, found := a[appName] ; !found {
		return nil, fmt.Errorf("Application '%s' not defined.", appName)
	} else {
		return v, nil
	}
}
