package forjfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApps(t *testing.T) {
	assert := assert.New(t)

	var apps AppsStruct

	/*************************************/
	testCase := "when apps is nil"

	app, err := apps.Found("test")
	assert.Nilf(app, "Expect Found function to return nothing %s", testCase)
	assert.EqualErrorf(err, "Application 'test' not defined.", "Expect an error %s", testCase)

	/*************************************/
	testCase = "when apps is initialized"

	apps = make(AppsStruct)

	app = new(AppStruct)
	app.Set("source1", "name", "name")

	apps["name"] = app

	app = new(AppStruct)

	app.Set("source1", "name", "app_name")

	apps["app_name"] = app

	appsMerged := make(AppsStruct)

	assert.Equalf(appsMerged, appsMerged.mergeFrom(nil), "Expect return appsMerged returned when from is nil %s", testCase)
	assert.Equalf(appsMerged, appsMerged.mergeFrom(apps), "Expect return appsMerged returned %s", testCase)
	assert.Equalf(len(apps), len(appsMerged), "Expect apps to be merged %s", testCase)

	var appsMergedNil AppsStruct
	appsMergedNil = appsMergedNil.mergeFrom(apps)
	assert.NotNilf(appsMergedNil, "Expect return new AppsStruct returned %s", testCase)
	assert.Equalf(len(apps), len(appsMergedNil), "Expect apps to be merged %s", testCase)

	testCase = "when appsMerged already initialized with existing elements."

	app = new(AppStruct)

	app.Set("source1", "name", "app_name2")

	apps["app_name2"] = app

	delete(apps, "app_name")

	assert.Equalf(appsMerged, appsMerged.mergeFrom(apps), "Expect return appsMerged returned %s", testCase)
	assert.Equalf(len(apps) + 1, len(appsMerged), "Expect apps to be merged %s", testCase)

	app, err = appsMerged.Found("name")
	assert.Equal("name", app.Name(), "Expect Found function to return the 'name' app %s", testCase)
	assert.NoErrorf(err, "Expect no error %s", testCase)

}
