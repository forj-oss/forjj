package main

import "strconv"

// Simple function to convert a dynamic type to bool
// it returns false by default except if the internal type is:
// - bool. value as is
// - string: call https://golang.org/pkg/strconv/#ParseBool
//
func to_bool(v interface{}) (bool) {
 switch v.(type) {
   case bool:
     return v.(bool)
   case string:
     s := v.(string)
     if b, err := strconv.ParseBool(s) ; err == nil { return b }
     return false
 }
 return false
}

// simply extract string from the dynamic type
// otherwise the returned string is empty.
func to_string(v interface{}) (result string) {
 switch v.(type) {
   case string:
     return v.(string)
 }
 return
}
