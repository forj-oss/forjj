package utils

import (
	"fmt"
)

// EvalSizes stores max for each elements evaluated.
type EvalValues struct {
	values   []int
	counts   []int
	evalFunc []func(int, int) int
}

// NewEvalValues creates an EvalValues object
//
// index : Number of evaluated values
func NewEvalValues(index int) (ret *EvalValues) {
	if index <= 0 {
		return
	}
	ret = new(EvalValues)
	ret.values = make([]int, index)
	ret.counts = make([]int, index)
	ret.evalFunc = make([]func(int, int) int, index)
	for index := range ret.evalFunc {
		ret.evalFunc[index] = Max
	}
	return
}

// Max is the default function to evaluate values and result as Max
func Max(ref, value int) int {
	if ref > value {
		return ref
	}
	return value
}

// Max is the default function to evaluate values and result as Max
func Cumulative(ref, value int) int {
	return value + ref
}

// Max is the default function to evaluate values and result as Max
func Min(ref, value int) int {
	if ref > value {
		return value
	}
	return ref
}

// Eval is used to evaluate an item value with the select eval function
//
func (m *EvalValues) Eval(index, value int) {
	if m == nil {
		return
	}
	if index < 0 || index > len(m.values) {
		return
	}

	m.values[index] = m.evalFunc[index](m.values[index], value)
	m.counts[index]++
}

// ValueOf return the evaluated indexed value after multiple EvalValues.Eval().
func (m *EvalValues) ValueOf(index int) (ret int) {
	if m == nil {
		return
	}
	if index < 0 || index > len(m.values) {
		return
	}

	return m.values[index]
}

// CountOf returns how many evaluation were executed on an indexed item
func (m *EvalValues) CountOf(index int) (ret int) {
	if m == nil {
		return
	}
	if index < 0 || index > len(m.values) {
		return
	}

	return m.counts[index]
}

// EvalFunc change the default Eval method to your Method
//
// Default is utils.Max
//
// Some predefined Eval functions:
//
// - utils.Cumulative
// - utils.Max
// - utils.Min
func (m *EvalValues) EvalFunc(index int, evalFunc func(int, int) int) {
	if m == nil {
		return
	}
	if index < 0 || index > len(m.values) {
		return
	}

	m.evalFunc[index] = evalFunc
}

// PrintfFormat format an Printf format with indexed ValueOf()
// The format is interpreted by Sprintf
//
// Ex: "%%-%ds" could be evaluated by "%-10s"
func (m *EvalValues) PrintfFormat(format string, indexes ...int) string {
	values := make([]interface{}, len(indexes))
	for iterator, index := range indexes {
		values[iterator] = m.ValueOf(index)
	}
	return fmt.Sprintf(format, values...)
}
