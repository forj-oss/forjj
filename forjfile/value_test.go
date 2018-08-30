package forjfile

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	assert := assert.New(t)

	value := ForjValue{}

	assert.Equal(false, value.SetDefault(""), "It should return false when default value is empty and not changed.")
	assert.Equal("", value.default_value, "Default value should be empty")

	assert.Equal(true, value.SetDefault("default"), "It should return true when default value is set")
	assert.Equal("default", value.default_value, "Default value should be empty")

	assert.Equal(false, value.Set(""), "It should return false when value to set is empty and not changed.")
	assert.Equal("", value.value, "Value should be empty")

	assert.Equal("default", value.Get(), "It should return the default value")
	assert.Equal(true, value.Set("value"), "It should consider value to be updated when set it.")
	assert.Equal(false, value.Set("value"), "It should consider value not to be updated when set it.")
	assert.Equal(false, value.IsDefault(), "It should return false to IsDefault.")
	assert.Equal("value", value.Get(), "It should return the value")
	assert.Equal(true, value.Set(""), "It should return true when value to set is empty and changed.")
	assert.Equal("default", value.Get(), "It should return the default value, when value is empty")
	assert.Equal(true, value.IsDefault(), "it should return true if is default value")

	value.Set("blabla")

	value.Clean("")

	assert.Equal("", value.default_value, "It should be cleaned up.")
	assert.Equal("", value.value, "It should be cleaned up.")

	yml, err := yaml.Marshal(value)
	assert.NoError(err, "No error reported, even empty.")
	value2 := ForjValue{}
	err = yaml.Unmarshal(yml, &value2)
	assert.NoError(err, "No error reported, even empty.")
	assert.Empty(value.Get(), "It should return an empty string. no default saved.")

	value.Set("value")
	yml, err = yaml.Marshal(value)
	assert.NoError(err, "No error reported.")
	err = yaml.Unmarshal(yml, &value2)
	assert.NoError(err, "No error reported.")
	assert.Equal("value", value.Get(), "It should return the value saved/restored.")
}
