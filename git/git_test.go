package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func logOutTest(string) {

}


func TestSetLogFunc(t *testing.T) {
	assert := assert.New(t)

	t.Log("Expect Log Func to properly store the logFunc")

	assert.NotNil(logFunc, "Expected logFunc to be set. Is nil.")
	assert.Equal(fmt.Sprintf("%v", logFunc), fmt.Sprintf("%v", logOut), "Expected default function to be registered by default.")

	SetLogFunc(logOutTest)

	assert.NotNil(logFunc, "Expected logFunc to be set. Is nil.")
	assert.NotEqual(fmt.Sprintf("%v", logFunc), fmt.Sprintf("%v", logOut), "Expected default function to be UNregistered.")
	assert.Equal(fmt.Sprintf("%v", logFunc), fmt.Sprintf("%v", logOutTest), "Expected new function to be registered.")

}
