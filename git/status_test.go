package git

import (
	"testing"
)

func TestAdd(t *testing.T) {
	t.Log("Expecting add to add files properly.")
	g := make(gitFiles)

	// Run the function
	t.Log("Running g.add(\"A\", \"aFile\")...")
	g.add("A", "aFile")

	// Test the result
	if v := len(g); v != 1 {
		t.Errorf("Expected gitFiles to contains 1 element. Got %d.", v)
	}
	if v, found := g["A"]; !found {
		t.Errorf("Expected gitFiles to contains the 'A' element. Not found.")
	} else if v == nil {
		t.Errorf("Expected gitFiles to contains a valid array attache to the 'A' element. Array not defined (nil)")
	} else if v2 := len(v); v2 != 1 {
		t.Errorf("Expected gitFiles to contains 1 element as 'A'. Got %d.", v2)
	} else if v[0] != "aFile" {
		t.Errorf("Expected gitFiles to contains file 'aFile' as element in 'A'. Got %d.", v2)
	}

	// Run the function
	t.Log("Running g.add(\"A\", \"bFile\")...")
	g.add("A", "bFile")

	// Test the result
	if v := len(g); v != 1 {
		t.Errorf("Expected gitFiles to contains 1 element. Got %d.", v)
	}
	if v, found := g["A"]; !found {
		t.Errorf("Expected gitFiles to contains the 'A' element. Not found.")
	} else if v == nil {
		t.Errorf("Expected gitFiles to contains a valid array attache to the 'A' element. Array not defined (nil)")
	} else if v2 := len(v); v2 != 2 {
		t.Errorf("Expected gitFiles to contains 1 element as 'A'. Got %d.", v2)
	} else if v[0] != "aFile" {
		t.Errorf("Expected gitFiles to contains file 'aFile' as element in 'A'. Got %d.", v2)
	} else if v[1] != "bFile" {
		t.Errorf("Expected gitFiles to contains file 'aFile' as element in 'A'. Got %d.", v2)
	}
}

func TestInit(t *testing.T) {
	t.Log("Expecting init .")
	g := make(gitFiles)

	// Run the function
	t.Log("Running g.init()...")
	g.init(false)

	// Test the result
	if v := len(g); v != 3 {
		t.Errorf("Expected gitFiles to contains 1 element. Got %d.", v)
	}
	if _, found := g["A"]; !found {
		t.Errorf("Expected gitFiles to contains the 'A' element. Not found.")
	} else if _, found = g["M"]; !found {
		t.Errorf("Expected gitFiles to contains the 'M' element. Not found.")
	} else if _, found = g["D"]; !found {
		t.Errorf("Expected gitFiles to contains the 'D' element. Not found.")
	}
}
