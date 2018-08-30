package forjfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppModel(t *testing.T) {
	assert := assert.New(t)

	/*********************************/
	testCase := "when app is nil"

	var app *AppStruct
	model := app.Model()

	assert.Nilf(model.app, "Expect model app to be nil %s", testCase)
	assert.Emptyf(model.Get("blabla"), "Expect model to return empty string %s", testCase)

	/*********************************/
	testCase = "when app is created"
	app = NewAppStruct()
	model = app.Model()
	assert.NotNil(app, "Expect App object returned %s", testCase)
	assert.NotNil(model.app, "Expect App object set in model %s", testCase)

	app.set("source", "name", "test", (*ForjValue).Set)
	app.set("source", "field1", "value1", (*ForjValue).Set)

	assert.Equalf(app.Name(), model.Get("name"), "Expect Name to be extracted from the model %s", testCase)
	assert.Equalf("value1", model.Get("field1"), "Expect model to extract field1 %s", testCase)
	assert.Emptyf(model.Get("blabla"), "Expect model to return empty string for unknown field %s", testCase)
}
