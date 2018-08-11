package forjfile

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestValues(t *testing.T) {
	assert := assert.New(t)

	values := make(ForjValues)

	assert.Empty(values, "It should be empty.")
	value1 := ForjValue{}
	value1.SetDefault("def1")
	value1.Set("val1")
	values["key1"] = value1

	value1.Set("val2")
	values["key2"] = value1

	value1.Set("")
	values["key3"] = value1

	value1 = values["key1"]
	assert.Equal("val1", value1.Get(), "It should get value1")
	value1 = values["key2"]
	assert.Equal("val2", value1.Get(), "It should get value2")
	value1 = values["key3"]
	assert.Equal("def1", value1.Get(), "It should get default value")

	mapExtract := values.Map()
	assert.Equal("val1", mapExtract["key1"], "mapExtract should have val1 at key1")
	assert.Equal("val2", mapExtract["key2"], "mapExtract should have val2 at key2")
	assert.Equal("", mapExtract["key3"], "mapExtract should be empty at key3")

	yml, err := yaml.Marshal(values)
	assert.NoError(err, "It should marshal data successfully.")

	valuesCopy := make(ForjValues)

	err = yaml.Unmarshal(yml, valuesCopy)
	assert.NoError(err, "It should unmarshal data successfully.")

	value1 = valuesCopy["key1"]
	assert.Equal("val1", value1.Get(), "It should get value1")
	value1 = valuesCopy["key2"]
	assert.Equal("val2", value1.Get(), "It should get value2")
	value1 = valuesCopy["key3"]
	assert.Equal("", value1.Get(), "It should get empty")

	ForjValueSelectDefault = true
	mapExtract = values.Map()
	assert.Equal("val1", mapExtract["key1"], "mapExtract should have val1 at key1")
	assert.Equal("val2", mapExtract["key2"], "mapExtract should have val2 at key2")
	assert.Equal("def1", mapExtract["key3"], "mapExtract should be empty at key3")

	yml, err = yaml.Marshal(values)
	assert.NoError(err, "It should marshal data successfully.")

	err = yaml.Unmarshal(yml, valuesCopy)
	assert.NoError(err, "It should unmarshal data successfully.")

	value1 = valuesCopy["key1"]
	assert.Equal("val1", value1.Get(), "It should get value1")
	value1 = valuesCopy["key2"]
	assert.Equal("val2", value1.Get(), "It should get value2")
	value1 = valuesCopy["key3"]
	assert.Equal("def1", value1.Get(), "It should get empty")


}
