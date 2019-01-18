package git

import (
	"testing"
	"reflect"

	"github.com/stretchr/testify/assert"
)

func logOutTest(string) {}


func TestSetLogFunc(t *testing.T) {
	assert := assert.New(t)

	t.Log("Expect Log Func to properly store the logFunc")

	assert.NotNil(logFunc, "Expected logFunc to be set. Is nil.")
	sflogFunc := reflect.ValueOf(logFunc).Pointer()
    sflogOut := reflect.ValueOf(logOut).Pointer()

	assert.Equal(sflogFunc, sflogOut, "Expected default function to be registered by default.")

	SetLogFunc(logOutTest)

	sflogFunc = reflect.ValueOf(logFunc).Pointer()
	sflogOutTest := reflect.ValueOf(logOutTest).Pointer()
	assert.NotNil(logFunc, "Expected logFunc to be set. Is nil.")
	assert.NotEqual(sflogFunc, sflogOut, "Expected default function to be UNregistered.")
	assert.Equal(sflogFunc, sflogOutTest, "Expected new function to be registered.")

}
