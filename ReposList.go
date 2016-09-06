package main

import (
    "github.com/alecthomas/kingpin"
    "regexp"
    "fmt"
    "strings"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
)

type ReposList struct {
    Repos map[string]*RepoStruct
}

// Syntax supported:
// --add-repo [<instance>/]<RepoName>[:<Flow>[:<RepoTemplate>[:<RepoTitle>]]][,...]
// --repo [<instance>/]<RepoName>[:<Flow>[:<RepoTemplate>[:<RepoTitle>]]][,...]

func (d *ReposList)Set(value string) error {
    if found, _ := regexp.MatchString(`[a-z0-9_:/-]+(,[a-z0-9_:/-]+)*`, value) ; !found {
        return fmt.Errorf("%s is an invalid list of Repository string. REPOS must be formated as '<REPO>[,<REPO>[...]] where REPO is formated as '[<instance>/]<RepoName>[:<FlowName>[:<RepoTemplate>[:<RepoTitle>]]]' all lower case. if no Flow is set, it will use the default one (--default-flow) or 'none'", value)
    }
    for _, v := range strings.Split(value, ",") {
        if err := d.Add(v) ; err != nil {
            return err
        }
    }
    return nil
}

func (d *ReposList)Add(value string) error {
    t, _ := regexp.Compile(`(([a-z]+[a-z0-9_-]*)/)?([a-z]+[a-z0-9_-]*)(:([a-z]+[a-z0-9_-]*)(:([a-z]+[a-z0-9_-]*)(:([a-z]+[a-z0-9_-]*))?)?)?`)
    res := t.FindStringSubmatch(value)
    if res == nil {
        return fmt.Errorf("%s is an invalid Repository. REPO must be formated as '[<instance>/]<RepoName>[:<FlowName>[:<RepoTemplate>[:<RepoTitle>]]]' all lower case. if no Flow is set, it will use the default one (--default-flow) or 'none'", value)
    }

    if d.Repos == nil {
        d.Repos = make(map[string]*RepoStruct)
    }
    r := &RepoStruct{
        Flow: res[5],
        Title: res[7],
        Templates:  make([]string,0,1),
        Groups : make(map[string]string),
        Users : make(map[string]string),
        Instance : res[1],
    }
    if res[9] != "" {
        r.Templates = append(r.Templates, res[7])
    }
    d.Repos[res[3]] = r
    gotrace.Trace("Repo added %s", value)
    return nil
}

// FIXME: kingpin is having trouble in the context case, where several --apps set, with some flags in between, is ignoring seconds and next --apps flags values. Workaround is to have them all followed or use the --apps APP[,APP ...] format.
func (d *ReposList)IsCumulative() bool {
    return true
}

func (d *ReposList)String() string {
    list := make([]string,0 , 2)

    for k, v := range d.Repos {
        list = append(list, fmt.Sprintf("%s:%s", k, v))
    }
    return strings.Join(list, ",")
}

// Set flag value type and assign it to the actionOpt struct
func SetReposListFlag(a *ActionOpts, f *kingpin.FlagClause) (*kingpin.FlagClause) {
    r := new(ReposList)
    f.SetValue(r)
    a.repoList = r
    return f
}
