package main

import (
    "gopkg.in/alecthomas/kingpin.v2"
    "regexp"
    "fmt"
)

type DriversList struct {
    list map[string]DriverDef
}

type DriverDef struct {
    Name string      // Name of the driver
    Type string      // Driver type
    Instance string  // Instance Name
}

func (d *DriversList)Set(value string) error {
    t, _ := regexp.Compile(`([a-z]+[a-z0-9_-]*):([a-z]+[a-z0-9_-]*)(:([a-z]+[a-z0-9_-]*))?`)
    res := t.FindStringSubmatch(value)
    if res == nil {
        return fmt.Errorf("%s is an invalid application driver. Must be formated as '<type>:<app>[:<Name>]' all lower case. if <Name> is missed, <Name> will be set to <app>", value)
    }

    dd := DriverDef{Type: res[1], Name: res[2]}

    instance := res[4]

    if instance == "" {
        instance = res[2]
    }
    dd.Instance = instance
    d.list[instance] = dd
    return nil
}

func (d *DriversList)IsCumulative() bool {
    return true
}

func (d *DriversList)String() string {
    return ""
}

func SetDriversListFlag(f *kingpin.FlagClause) (*kingpin.FlagClause) {
    f.SetValue(new(DriversList))
    return f
}

func (values *DriversList) GetDriversFromContext(context *kingpin.ParseContext, f *kingpin.FlagClause) (found bool) {
    for _, element := range context.Elements {
        if flag, ok := element.Clause.(*kingpin.FlagClause); ok && flag == f {
            values.Set(*element.Value)
            found = true
        }
    }
    return
}

