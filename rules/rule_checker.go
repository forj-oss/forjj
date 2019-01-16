package rules

import (
	"errors"
	"fmt"
	"regexp"
)

// rules package defines a rule checker system.
// It checks the rule for a key and a value. The value is returned by a get function.
//
// a rule is a string formatted with following syntax and meaning:
// - '<key>:<value>' or '<key>=<value>' - True if key has value equal to <value>. '<key>:<value>' is kept for compatibility but is obsolete.
// - '<key>!=<value>' - True if key has a value NOT equal to <value>
// - '<key>=/<regexp>/' - True if key has value respecting <regexp>.
// - '<key>!=/<regexp>/' - True if key has a value NOT respecting <regexp>

// RuleChecker is the core Object to manage a rule check.
type RuleChecker struct {
	ruleRE         *regexp.Regexp
	ruleToCheck    []string
	key, op, value string
}

// NewRuleChecker returns a RuleChecker object
func NewRuleChecker() (ret *RuleChecker) {
	ret = new(RuleChecker)
	ret.ruleRE, _ = regexp.Compile(`^(.*?)(:|!?=(/)?)(.*?)(/)?$`)
	return
}

// Validate check if the rule string is valid and can be used to check.
// This function must be called before Check()
func (r *RuleChecker) Validate(rule string) (key, op, value string, err error) {
	if r == nil {
		err = errors.New("Invalid RuleChecker object")
		return
	}

	r.key = ""
	r.op = ""
	r.value = ""

	r.ruleToCheck = r.ruleRE.FindStringSubmatch(rule)
	if r.ruleToCheck == nil {
		err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>[!]=<value>', <key>[!]=/<regExp>/", rule)
		return
	}

	if r.ruleToCheck[3] == "/" && r.ruleToCheck[5] != "/" {
		err = errors.New("RegExp format error. Missing leading '/'")
		return
	}

	r.key = r.ruleToCheck[1]
	r.op = r.ruleToCheck[2]
	r.value = r.ruleToCheck[4]

	if r.ruleToCheck[3] == "" {
		r.value += r.ruleToCheck[5]
	}

	return r.key, r.op, r.value, nil
}

// Check check if the key respect the value/regexp and returns true if confirmed.
//
// The string value is given by the call to the `get` function
func (r *RuleChecker) Check(get func(string) (string, bool)) (matched bool, err error) {
	if r == nil {
		return false, errors.New("Invalid RuleChecker object")
	}

	value, found := get(r.key)

	switch r.op {
	case "=", ":":
		if r.value == "*" {
			return found, nil
		}
		return found && r.value == value, nil
	case "!=":
		if r.value == "*" {
			return !found, nil
		}
		return r.value != value, nil
	case "=/":
		matched, err = regexp.MatchString(r.value, value)
		return
	case "!=/":
		matched, err = regexp.MatchString(r.value, value)
		if err != nil {
			return
		}
		return !matched, nil
	}

	return
}
