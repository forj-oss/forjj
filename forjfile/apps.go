package forjfile

import "fmt"

type AppsStruct map[string]*AppStruct

func (a AppsStruct) Found(appName string) (*AppStruct, error) {
	if v, found := a[appName]; !found {
		return nil, fmt.Errorf("Application '%s' not defined.", appName)
	} else {
		return v, nil
	}
}

func (a AppsStruct) mergeFrom(from AppsStruct) {
	for k, appFrom := range from {
		if app, found := a[k]; found {
			app.mergeFrom(appFrom)
		} else {
			
			a[k] = appFrom
		}
	}

}
