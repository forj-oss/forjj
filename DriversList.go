package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	"regexp"
	"strings"
)

type DriversList struct {
	list map[string]DriverDef
}

type DriverDef struct {
	Name     string // Name of the driver
	Type     string // Driver type
	Instance string // Instance Name
}

func (d *DriversList) Set(value string) error {
	if found, _ := regexp.MatchString(`[a-z0-9_:-]+(,[a-z0-9_:-]*)*`, value); !found {
		return fmt.Errorf("%s is an invalid list of applications string. APPS must be formated as '<APP>[,<APP>[...]] where APP is formated as <type>:<DriverName>[:<InstanceName>]' all lower case. if <Name> is missed, <Name> will be set to <app>", value)
	}
	for _, v := range strings.Split(value, ",") {
		if err := d.Add(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DriversList) Add(value string) error {
	t, _ := regexp.Compile(`([a-z]+[a-z0-9_-]*):([a-z]+[a-z0-9_-]*)(:([a-z]+[a-z0-9_-]*))?`)
	res := t.FindStringSubmatch(value)
	if res == nil {
		return fmt.Errorf("%s is an invalid application driver. APP must be formated as '<type>:<DriverName>[:<InstanceName>]' all lower case. if <Name> is missed, <Name> will be set to <app>", value)
	}

	dd := DriverDef{Type: res[1], Name: res[2]}

	instance := res[4]

	if instance == "" {
		instance = res[2]
	}
	dd.Instance = instance
	if d.list == nil {
		d.list = make(map[string]DriverDef)
	}
	d.list[instance] = dd
	gotrace.Trace("Driver added %s", value)
	return nil
}

// FIXME: kingpin is having trouble in the context case, where several --apps set, with some flags in between, is ignoring seconds and next --apps flags values. Workaround is to have them all followed or use the --apps APP[,APP ...] format.
func (d *DriversList) IsCumulative() bool {
	return true
}

func (d *DriversList) String() string {
	list := make([]string, 0, 2)

	for _, v := range d.list {
		var s string

		if v.Instance == v.Name {
			s = fmt.Sprintf("%s:%s", v.Type, v.Name)
		} else {
			s = fmt.Sprintf("%s:%s:%s", v.Type, v.Name, v.Instance)
		}
		list = append(list, s)
	}
	return strings.Join(list, ",")
}

func SetDriversListFlag(f *kingpin.FlagClause) *kingpin.FlagClause {
	f.SetValue(new(DriversList))
	return f
}

func (values *DriversList) GetDriversFromContext(context *kingpin.ParseContext, f *kingpin.FlagClause) (found bool) {
	for _, element := range context.Elements {
		if flag, ok := element.Clause.(*kingpin.FlagClause); ok && flag == f {
			values.Set(*element.Value)
			gotrace.Trace("Context Found --apps %s\n", *element.Value)
			found = true
		}
	}
	//return
}
