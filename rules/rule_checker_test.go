package rules

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuleChecker(t *testing.T) {
	assert := assert.New(t)

	t.Log("Expecting NewRuleChecker to properly create a RuleChecker object")

	rule := NewRuleChecker()

	assert.NotNil(rule, "Expect to get an object. Got nil.")
	if rule == nil {
		return
	}

	assert.NotNil(rule.ruleRE, "Expect to have a valid Rule regexp defined. Got nil")

}

func TestValidate(t *testing.T) {
	assert := assert.New(t)

	t.Log("Expecting RuleChecker.Validate to properly check rule given and prepare for Check.")

	ruleChecker := NewRuleChecker()

	testList := validateTestCases{
		validateTestCase{"blabla", "", "", "", false},
		validateTestCase{"blabla:", "blabla", ":", "", true},
		validateTestCase{"blabla=", "blabla", "=", "", true},
		validateTestCase{"blabla=test", "blabla", "=", "test", true},
		validateTestCase{"blabla!=test", "blabla", "!=", "test", true},
		validateTestCase{"blabla!:", "blabla!", ":", "", true},
		validateTestCase{"blabla=/test/", "blabla", "=/", "test", true},
		validateTestCase{"blabla!=/test/", "blabla", "!=/", "test", true},
		validateTestCase{"blabla=/test", "", "", "", false},
		validateTestCase{"blabla!=/test", "", "", "", false},
		validateTestCase{"blabla!:/test/", "blabla!", ":", "/test/", true},
		validateTestCase{"blabla:test=toto", "blabla", ":", "test=toto", true},
		validateTestCase{"blabla=test=toto", "blabla", "=", "test=toto", true},
		validateTestCase{"blabla=test=/toto/", "blabla", "=", "test=/toto/", true},
	}

	testList.assertValidateAll(assert, ruleChecker)
}

func TestCheck(t *testing.T) {
	assert := assert.New(t)

	t.Log("Expecting RuleChecker.Validate to properly check rule given and prepare for Check.")

	ruleChecker := NewRuleChecker()

	testList := checkTestCases{
		// rule, value from key,             found, match, no error
		checkTestCase{"test:value", "value", true, true, true},
		checkTestCase{"test:value2", "value", true, false, true},
		checkTestCase{"test:value", "", false, false, true},
		checkTestCase{"test!=value", "value", true, false, true},
		checkTestCase{"test!=value", "value2", true, true, true},
		checkTestCase{"test=/value/", "value2", true, true, true},
		checkTestCase{"test=/^value$/", "value2", true, false, true},
		checkTestCase{"test=/^value//", "value/", true, true, true},
		checkTestCase{"test=/^value/$/", "value/", true, true, true},
		checkTestCase{"test=/[value/", "value2", true, true, false},
		checkTestCase{"test=/value", "value2", true, true, false},
		checkTestCase{"test!=/[value/", "value2", true, true, false},
		checkTestCase{"test!=/value/", "value2", true, false, true},
		checkTestCase{"test=*", "value2", true, true, true},
		checkTestCase{"test!=*", "value2", true, false, true},
		checkTestCase{"test!=*", "", false, true, true},
	}

	testList.assertCheckAll(assert, ruleChecker)
}

/*******************************
        Test Suite code
********************************/

// Validate function test code

type validateTestCases []validateTestCase

func (tc validateTestCases) assertValidateAll(assert *assert.Assertions, rule *RuleChecker) {
	for _, testCase := range tc {
		testCase.assertValidate(assert, rule)
	}
}

type validateTestCase struct {
	rule           string
	key, op, value string
	noerror        bool
}

func (tc *validateTestCase) assertValidate(assert *assert.Assertions, rule *RuleChecker) {
	testCase := fmt.Sprintf("when rule is '%s'.", tc.rule)

	theKey, theOp, theValue, theErr := rule.Validate(tc.rule)

	if tc.noerror {
		assert.NoErrorf(theErr, "Expect no error %s", testCase)
		assert.Equalf(tc.key, theKey, "Expect key to be set properly %s", testCase)
		assert.Equalf(tc.op, theOp, "Expect op to be set properly %s", testCase)
		assert.Equalf(tc.value, theValue, "Expect value to be set properly %s", testCase)
		return
	}
	assert.Errorf(theErr, "Expect an error %s", testCase)
	assert.Emptyf(theKey, "Expect key to be empty %s", testCase)
	assert.Emptyf(theOp, "Expect op to be empty %s", testCase)
	assert.Emptyf(theValue, "Expect value to be empty %s", testCase)
}

// Validate function test code

type checkTestCases []checkTestCase

func (tc checkTestCases) assertCheckAll(assert *assert.Assertions, rule *RuleChecker) {
	for _, testCase := range tc {
		testCase.assertCheck(assert, rule)
	}
}

type checkTestCase struct {
	rule    string
	value   string
	found   bool
	matched bool
	noerror bool
}

func (tc *checkTestCase) assertCheck(assert *assert.Assertions, rule *RuleChecker) {
	testCase := fmt.Sprintf("when rule is '%s' and value=%s, found=%t.", tc.rule, tc.value, tc.found)

	if _, _, _, err := rule.Validate(tc.rule); !tc.noerror && err != nil {
		// expect an error and got an error. Normal case.
		return
	} else if err != nil {
		assert.NoErrorf(err, "Expect no error on validate %s", testCase)
	}

	var matched bool
	matched, err := rule.Check(func(string) (string, bool) {
		return tc.value, tc.found
	})

	if tc.noerror {
		assert.NoErrorf(err, "Expect no error %s", testCase)
		if tc.matched {
			assert.Truef(matched, "Expect rule to match %s", testCase)
		} else {
			assert.Falsef(matched, "Expect rule to not match %s", testCase)
		}
		return
	}
	assert.Errorf(err, "Expect an error %s", testCase)
}
