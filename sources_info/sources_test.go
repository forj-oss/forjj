package sourcesinfo

import (
	"testing"
)

func TestSet(t *testing.T) {
	t.Log("Expect set to initialize the sources struct.")

	var sources *Sources

	const (
		src1   = "src1"
		src2   = "src2"
		key1   = "key1"
		key2   = "key2"
		value1 = "value1"
		value2 = ""
	)
	// ------------ Run function to test
	// sources is nil
	sources = sources.Set(src1, key1, value1)

	// ------------ Test result
	if sources == nil {
		t.Error("Expect set to return a new object")
	} else if v1, f1 := sources.keys[key1]; !f1 {
		t.Error("Expect set to set a source key. Not found")
	} else if v1 != src1 {
		t.Errorf("Expect set to set a source value = '%s'. got '%s'", src1, v1)
	}

	// ------------ update context
	// Change source information
	sources = sources.Set(src2, key1, value1)
	// ------------ Test result
	if sources == nil {
		t.Error("Expect set to return a new object")
	} else if v1, f1 := sources.keys[key1]; !f1 {
		t.Error("Expect set to set a source key. Not found")
	} else if v1 != src2 {
		t.Errorf("Expect set to set a source value = '%s'. got '%s'", src2, v1)
	}

	// ------------ update context
	// add a new key, with a different source
	sources = sources.Set(src1, key2, value1)
	// ------------ Test result
	if sources == nil {
		t.Error("Expect set to return a new object")
	} else if v1, f1 := sources.keys[key1]; !f1 {
		t.Errorf("Expect set to set a source key '%s'. Not found", key1)
	} else if v1 != src2 {
		t.Errorf("Expect set to set a source value = '%s'. got '%s'", src2, v1)
	} else if v2, f2 := sources.keys[key2]; !f2 {
		t.Errorf("Expect set to set a source key '%s'. Not found", key2)
	} else if v2 != src1 {
		t.Errorf("Expect set to set a source value = '%s'. got '%s'", src2, v1)
	}
	// ------------ update context
	// add a new key wth value empty
	sources = sources.Set(src1, key2, value2)
	// ------------ Test result
	if sources == nil {
		t.Error("Expect set to return a new object")
	} else if v1, f1 := sources.keys[key1]; !f1 {
		t.Errorf("Expect set to set a source key '%s'. Not found", key1)
	} else if v1 != src2 {
		t.Errorf("Expect set to set a source value = '%s'. got '%s'", src2, v1)
	} else if _, f2 := sources.keys[key2]; f2 {
		t.Errorf("Expect set to unset a source key '%s'. found it", key2)
	}
}

func TestGet(t *testing.T) {
	t.Log("Expect get to retrun the source information.")

	var sources *Sources

	const (
		src1   = "src1"
		src2   = "src2"
		key1   = "key1"
		key2   = "key2"
		key3   = "key3"
		value1 = "value1"
		value2 = ""
	)
	// ------------ Run function to test
	// sources is nil
	ret := sources.Get(key1)

	// ------------ Test result

	if ret != "" {
		t.Errorf("Expect get to return an empty string. Got '%s'", ret)
	}

	// ------------ update context
	sources = sources.Set(src1, key1, value1)
	sources = sources.Set(src2, key2, value1)

	// ------------ Run function to test
	// check source information returned
	ret = sources.Get(key1)
	ret2 := sources.Get(key2)
	ret3 := sources.Get(key3)

	// ------------ Test result
	if ret != src1 {
		t.Errorf("Expect get to return '%s'. Got '%s'", src1, ret)
	} else if ret2 != src2 {
		t.Errorf("Expect get to return '%s'. Got '%s'", src1, ret2)
	} else if ret3 == src2 {
		t.Errorf("Expect get to return ''. Got '%s'", ret3)
	}
}
