package forjfile

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	assert := assert.New(t)

	expectedFlags := []string{appDriver, appName, appType}

	/*********************************/
	testCase := "when app is nil"

	var app *AppStruct

	assert.Nilf(app.Flags(), "Expect Flags to be nil %s", testCase)
	assert.NoErrorf(app.UnmarshalYAML(func(interface{}) error { return nil }), "Expect unmarshal to return no error %s", testCase)
	v, err := app.MarshalYAML()
	assert.NoErrorf(err, "Expect no error Marshaling %s", testCase)
	assert.Nilf(v, "Expect no interface %s", testCase)
	assert.Emptyf(app.Name(), "Expect name emty %s", testCase)

	value, found, source := app.Get("blabla")
	assert.Emptyf(value, "Expect value to be empty %s", testCase)
	assert.Falsef(found, "Expect found to be false %s", testCase)
	assert.Emptyf(source, "Expect source to be empty %s", testCase)

	assert.Falsef(app.SetHandler("blabla", nil, nil), "Expect SetHandler to return false %s", testCase)

	assert.Falsef(app.Set("blabla", "blabla", "blabla"), "Expect Set to return false %s", testCase)

	assert.Falsef(app.set("blabla", "blabla", "blabla", nil), "Expect Set to return false %s", testCase)

	/*********************************/
	testCase = "when app is created"
	app = NewAppStruct()
	assert.NotNilf(app, "Expect App object returned %s", testCase)

	assert.ElementsMatchf(expectedFlags, app.Flags(), "Expect minimal flags %s", testCase)

	/*********************************/
	testCase = "when app is initialized with standard fields and 1 extra field"

	app.Set("source", appName, "name")
	app.Set("source", appDriver, "driver")
	app.Set("source", appType, "type")
	app.Set("source", "field1", "value1")

	assert.Equalf("name", app.Name(), "Expect name to be set to '%s' %s", "name", testCase)
	assert.Equalf("driver", app.Driver, "Expect driver to be set to '%s' %s", "driver", testCase)
	assert.Equalf("type", app.Type, "Expect type to be set to '%s' %s", "type", testCase)

	value, found, source = app.Get("name")
	assert.Truef(found, "Expect name is found %s", testCase)
	assert.Equalf("source", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("name", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app.Get("type")
	assert.Truef(found, "Expect name is found %s", testCase)
	assert.Equalf("source", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("type", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app.Get("driver")
	assert.Truef(found, "Expect driver is found %s", testCase)
	assert.Equalf("source", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("driver", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app.Get("field1")
	assert.Truef(found, "Expect field1 is found %s", testCase)
	assert.Equalf("source", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("value1", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app.Get("field2")
	assert.Falsef(found, "Expect field2 is not found %s", testCase)
	assert.Emptyf(source, "Expect source to be set properly %s", testCase)
	assert.Emptyf(value.GetString(), "Expect name to be properly set %s", testCase)

	/*********************************/
	testCase = "when app is initialized with SetHandler and ForjValue.SetDefault"

	updated := app.SetHandler("source1", func(field string) (val string, found bool) {
		vals := map[string]string{
			"field1": "Defaultvalue2",
			"field2": "Defaultvalue3",
		}

		val, found = vals[field]
		return
	}, (*ForjValue).SetDefault, "field1", "field2")

	assert.Truef(updated, "Expect SetHandler to return true %s", testCase)

	value, found, source = app.Get("field1")
	assert.Truef(found, "Expect field1 is found %s", testCase)
	assert.Equalf("source1", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("value1", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app.Get("field2")
	assert.Truef(found, "Expect field2 is found %s", testCase)
	assert.Equalf("source1", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("Defaultvalue3", value.GetString(), "Expect name to be properly set %s", testCase)

	/*********************************/
	testCase = "when app is initialized with SetHandler and ForjValue.Set"

	updated = app.SetHandler("source2", func(field string) (val string, found bool) {
		vals := map[string]string{
			"field3": "value4",
			"field2": "value3",
		}

		val, found = vals[field]
		return
	}, (*ForjValue).Set, "field3", "field2")

	assert.Truef(updated, "Expect SetHandler to return true %s", testCase)

	value, found, source = app.Get("field3")
	assert.Truef(found, "Expect field3 is found %s", testCase)
	assert.Equalf("source2", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("value4", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app.Get("field2")
	assert.Truef(found, "Expect field2 is found %s", testCase)
	assert.Equalf("source2", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("value3", value.GetString(), "Expect name to be properly set %s", testCase)

	expectedFlags = []string{appDriver, appName, appType, "field1", "field2", "field3"}
	assert.ElementsMatchf(expectedFlags, app.Flags(), "Expect all flags %s", testCase)

	/*********************************/
	testCase = "when app is initialized with SetHandler and ForjValue.Clean, and value empty"

	updated = app.SetHandler("source2", func(field string) (val string, found bool) {
		vals := map[string]string{
			"field2": "",
		}

		val, found = vals[field]
		return
	}, (*ForjValue).Clean, "field2")

	assert.Truef(updated, "Expect SetHandler to return true %s", testCase)

	value, found, source = app.Get("field2")
	assert.Falsef(found, "Expect field2 to be not found %s", testCase)
	assert.Emptyf(source, "Expect source to be empty %s", testCase)
	assert.Emptyf(value.GetString(), "Expect name to be empty %s", testCase)

	expectedFlags = []string{appDriver, appName, appType, "field1", "field3"}
	assert.ElementsMatchf(expectedFlags, app.Flags(), "Expect all flags minus removed %s", testCase)

	/*********************************/
	testCase = "when app2 is merged but nil."

	var app2 *AppStruct

	app2.mergeFrom(app)

	assert.Nilf(app2, "Expect app2 still nil.", testCase)
	assert.NotNilf(app, "Expect app not nil %s", testCase)
	assert.ElementsMatchf(expectedFlags, app.Flags(), "Expect no flags changed %s", testCase)

	/*********************************/
	testCase = "when app2 is merged."

	app2 = NewAppStruct()

	assert.NotNilf(app2, "Expect app2 not nil %s", testCase)

	updated = app2.SetHandler("source3", func(field string) (val string, found bool) {
		vals := map[string]string{
			"name":   "name2",
			"type":   "type2",
			"field4": "value4",
		}

		val, found = vals[field]
		return
	}, (*ForjValue).Set, "name", "type", "field4")

	assert.Truef(updated, "Expect SetHandler to return true %s", testCase)

	app2.mergeFrom(app)

	expectedFlags = []string{appDriver, appName, appType, "field1", "field3"}
	assert.ElementsMatchf(expectedFlags, app.Flags(), "Expect all flags minus removed %s", testCase)

	expectedFlags = []string{appDriver, appName, appType, "field1", "field3", "field4"}
	assert.ElementsMatchf(expectedFlags, app2.Flags(), "Expect all flags minus removed %s", testCase)

	value, found, source = app2.Get("field4")
	assert.Truef(found, "Expect field4 is found %s", testCase)
	assert.Equalf("source3", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("value4", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app2.Get("name")
	assert.Truef(found, "Expect name is found %s", testCase)
	assert.Equalf("source", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("name", value.GetString(), "Expect name to be properly set %s", testCase)

	value, found, source = app2.Get("type")
	assert.Truef(found, "Expect type is found %s", testCase)
	assert.Equalf("source", source, "Expect source to be set properly %s", testCase)
	assert.Equalf("type", value.GetString(), "Expect name to be properly set %s", testCase)

	/*********************************/
	testCase = "when app is marshaled and unmarshaled in app2"

	yml, err := yaml.Marshal(app)

	assert.NoErrorf(err, "Expect no error %s", testCase)
	assert.NotNilf(yml, "Expect yaml data generated %s", testCase)

	err = yaml.Unmarshal(yml, app2)

	assert.NoErrorf(err, "Expect no error %s", testCase)

	assert.ElementsMatchf(app.Flags(), app2.Flags(), "Expect no flags changed %s", testCase)

}

func TestAppModel(t *testing.T) {
	assert := assert.New(t)

	/*********************************/
	testCase := "when app is nil"

	var app *AppStruct
	model := app.Model()

	assert.Nilf(model.app, "Expect model app to be nil %s", testCase)

	/*********************************/
	testCase = "when app is created"
	app = NewAppStruct()
	model = app.Model()
	assert.NotNil(app, "Expect App object returned %s", testCase)
	assert.NotNil(model.app, "Expect App object set in model %s", testCase)

}
