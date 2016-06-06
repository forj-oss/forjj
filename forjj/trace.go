package main

import (
        "fmt"
        "runtime"
)

var debug *bool

func trace(s string, a ...interface{}) {
 if ! *debug { return }

 pc := make([]uintptr, 10)  // at least 1 entry needed
 runtime.Callers(2, pc)
 f := runtime.FuncForPC(pc[0])
 fmt.Printf(fmt.Sprintf("DEBUG %s: %s", f.Name(), s), a ...)
}

// Display internal structs

func (o *DriverCmdOptions)String() (s string) {
 s = "flags:"
 for k,v := range o.flags {
   s = fmt.Sprintf("%s\n- %s => %s", k, v)
 }

 s = fmt.Sprintf("%s\nargs:", s)
 for k,v := range o.args {
   s = fmt.Sprintf("%s\n- %s => %s", k, v)
 }
 return
}

func (o *DriverOptions)String() (s string) {
 s = fmt.Sprintf("Driver name: %s\ntype: %s\n", o.name, o.driver_type)
 for k,v := range o.cmds {
   s = fmt.Sprintf("%sDriverCmdOptions: %s:\n%s", k, v.String())
 }
 return
}

