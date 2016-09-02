package main

import (
    "github.com/alecthomas/kingpin"
    "regexp"
    "fmt"
    "strings"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
)

type ReposList struct {
    list map[string]RepoStruct
}

func (d *ReposList)Set(value string) error {
    if found, _ := regexp.MatchString(`[a-z0-9_:-]+(,[a-z0-9_-]+)*`, value) ; !found {
        return fmt.Errorf("%s is an invalid list of Repository string. REPOS must be formated as '<REPO>[,<REPO>[...]] where REPO is formated as '<RepoName>[:<FlowName>]' all lower case. if no Flow is set, it will use the default one (--default-flow) or 'none'", value)
    }
    for _, v := range strings.Split(value, ",") {
        if err := d.Add(v) ; err != nil {
            return err
        }
    }
    return nil
}

func (d *ReposList)Add(value string) error {
    t, _ := regexp.Compile(`([a-z]+[a-z0-9_-]*)(:([a-z]+[a-z0-9_-]*))?(:([a-z0-9_-]*))?`)
    res := t.FindStringSubmatch(value)
    if res == nil {
        return fmt.Errorf("%s is an invalid Repository. REPO must be formated as '<RepoName>[:<FlowName>]' all lower case. if no Flow is set, it will use the default one (--default-flow) or 'none'", value)
    }

    if d.list == nil {
        d.list = make(map[string]RepoStruct)
    }
    d.list[res[1]] = RepoStruct{
        Flow: res[3],
    }
    gotrace.Trace("Driver added %s", value)
    return nil
}

// FIXME: kingpin is having trouble in the context case, where several --apps set, with some flags in between, is ignoring seconds and next --apps flags values. Workaround is to have them all followed or use the --apps APP[,APP ...] format.
func (d *ReposList)IsCumulative() bool {
    return true
}

func (d *ReposList)String() string {
    list := make([]string,0 , 2)

    for k, v := range d.list {
        list = append(list, fmt.Sprintf("%s:%s", k, v))
    }
    return strings.Join(list, ",")
}

func SetReposListFlag(f *kingpin.FlagClause) (*kingpin.FlagClause) {
    f.SetValue(new(ReposList))
    return f
}

func (values *ReposList) GetDriversFromContext(context *kingpin.ParseContext, f *kingpin.FlagClause) (found bool) {
    for _, element := range context.Elements {
        if flag, ok := element.Clause.(*kingpin.FlagClause); ok && flag == f {
            values.Set(*element.Value)
            gotrace.Trace("Context Found --repos %s\n", *element.Value)
            found = true
        }
    }
    return
}

