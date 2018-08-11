package forjfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	assert := assert.New(t)

	user := new(UserStruct)

	expectedFlags := []string{userRole}

	assert.ElementsMatch(expectedFlags, user.Flags(), "Initial flags should be limited to userRole.")

	user.Set("source", userRole, "role")
	assert.ElementsMatch(expectedFlags, user.Flags(), "Initial flags should be limited to userRole.")
	value, found, source := user.Get(userRole)
	assert.True(found, "It should find it.")
	assert.Equal("source", source, "Source should be returned as expected.")
	assert.Equal("role", value.GetString(), "We should get the role.")

	const (
		newflag = "newflag"
		value1  = "value1"
		value2  = "value2"
		value3  = "value3"
	)
	user.Set("source2", newflag, value2)
	expectedFlags = []string{userRole, newflag}
	assert.ElementsMatch(expectedFlags, user.Flags(), "All existing flags must be identified.")

	value, found, source = user.Get(userRole)
	assert.Truef(found, "It should find %s.", userRole)
	assert.Equal("source", source, "Source should be returned as expected.")
	assert.Equal("role", value.GetString(), "We should get the role.")

	value, found, source = user.Get(newflag)
	assert.Truef(found, "It should find %s.", newflag)
	assert.Equal("source2", source, "Source should be returned as expected.")
	assert.Equalf(value2, value.GetString(), "We should get the value '%s'.", value2)

	user.Set("source3", newflag, value3)
	assert.ElementsMatch(expectedFlags, user.Flags(), "All existing flags must be identified.")

	value, found, source = user.Get(userRole)
	assert.Truef(found, "It should find %s.", userRole)
	assert.Equal("source", source, "Source should be returned as expected.")
	assert.Equal("role", value.GetString(), "We should get the role.")

	value, found, source = user.Get(newflag)
	assert.Truef(found, "It should find %s.", newflag)
	assert.Equal("source3", source, "Source should be returned as expected.")
	assert.Equalf(value3, value.GetString(), "We should get the value '%s'.", value3)

	userToMerge := new(UserStruct)
	const (
		value4   = "value4"
		newflag2 = "newflag2"
		newflag3 = "newflag3"
		value5   = "value5"
	)

	userToMerge.Set("source4", userRole, "role2")
	userToMerge.Set("source4", newflag, value4)
	userToMerge.Set("source4", newflag2, value5)
	user.Set("source", newflag3, value5)

	user.mergeFrom(userToMerge)

	value, found, source = user.Get(userRole)
	assert.Truef(found, "It should find %s.", userRole)
	assert.Equal("source4", source, "Source should be returned as expected.")
	assert.Equal("role2", value.GetString(), "We should get the role.")

	value, found, source = user.Get(newflag)
	assert.Truef(found, "It should find %s.", newflag)
	assert.Equal("source4", source, "Source should be returned as expected.")
	assert.Equalf(value4, value.GetString(), "We should get the value '%s'.", value4)

	value, found, source = user.Get(newflag2)
	assert.Truef(found, "It should find %s.", newflag2)
	assert.Equal("source4", source, "Source should be returned as expected.")
	assert.Equalf(value5, value.GetString(), "We should get the value '%s'.", value5)

	value, found, source = user.Get(newflag3)
	assert.Truef(found, "It should find %s.", newflag2)
	assert.Equal("source", source, "Source should be returned as expected.")
	assert.Equalf(value5, value.GetString(), "We should get the value '%s'.", value5)

	const (
		key1 = "key1"
		key2 = "key2"
	)
	user.SetHandler("source5", func(field string) (ret string, found bool) {
		val := map[string]string{
			key1: value1,
			key2: value2,
		}
		ret, found = val[field]
		return
	}, key1, key2)

	value, found, source = user.Get(key1)
	assert.Truef(found, "It should find %s.", key1)
	assert.Equal("source5", source, "Source should be returned as expected.")
	assert.Equalf(value1, value.GetString(), "We should get the value '%s'.", value1)

	value, found, source = user.Get(key2)
	assert.Truef(found, "It should find %s.", key2)
	assert.Equal("source5", source, "Source should be returned as expected.")
	assert.Equalf(value2, value.GetString(), "We should get the value '%s'.", value2)
}
