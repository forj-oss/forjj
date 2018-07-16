package forjfile

import (
	"forjj/sources_info"

	"github.com/forj-oss/goforjj"
)

type GroupsStruct map[string]*GroupStruct

func (g GroupsStruct) mergeFrom(from GroupsStruct) GroupsStruct {
	if from == nil {
		return g
	}
	if g == nil {
		g = make(GroupsStruct)
	}
	for k, groupFrom := range from {
		if group, found := g[k]; found {
			group.mergeFrom(groupFrom)
		} else {
			g[k] = groupFrom
		}
	}
	return g
}

type GroupStruct struct {
	forge   *ForgeYaml
	Role    string            `yaml:",omitempty"`
	Members []string          `yaml:",omitempty"`
	More    map[string]string `yaml:",inline"`
	sources *sourcesinfo.Sources
}

const (
	groupRole    = "role"
	groupMembers = "members"
)

// TODO: Add struct unit tests

// Flags returns the list of keys of this object.
func (a *GroupStruct) Flags() (flags []string) {
	flags = make([]string, 2, 2+len(a.More))
	flags[0] = groupRole
	flags[1] = groupMembers
	for k := range a.More {
		flags = append(flags, k)
	}
	return
}

func (g *GroupStruct) Get(field string) (value *goforjj.ValueStruct, found bool, source string) {
	source = g.sources.Get(field)
	switch field {
	case "role":
		value, found = value.SetIfFound(g.Role, (g.Role != ""))
	case "members":
		value, found = value.SetIfFound(g.Members, (g.Members != nil && len(g.Members) > 0))
	default:
		v, f := g.More[field]
		value, found = value.SetIfFound(v, f)
	}
	return
}

func (g *GroupStruct) GetMembers() []string {
	return g.Members
}

func (g *GroupStruct) AddMembers(members ...string) (count int) {
	add_members := map[string]int{}
	for _, new_member := range members {
		if g.hasMember(new_member) >= 0 {
			add_members[new_member] = 0
		}
	}
	count = len(add_members)
	if count > 0 {
		g.forge.dirty()
	}
	new_members := make([]string, 0, len(add_members)+len(g.Members))
	copy(new_members, g.Members)
	copy(new_members, keys(add_members))
	g.Members = new_members
	return
}

func (g *GroupStruct) RemoveMembers(members ...string) (count int) {
	for _, new_member := range members {
		if index := g.hasMember(new_member); index >= 0 {
			g.Members = removeSliceString(g.Members, index)
			count++
		}
	}
	if count > 0 {
		g.forge.dirty()
	}
	return
}

func (g *GroupStruct) HasMember(member_exist string) bool {
	return (g.hasMember(member_exist) >= 0)
}

func (g *GroupStruct) hasMember(member_exist string) int {
	for index, member := range g.Members {
		if member == member_exist {
			return index
		}
	}
	return -1
}

func keys(mymap map[string]int) (keys []string) {
	keys = make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	return
}

func removeSliceString(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (g *GroupStruct) set_forge(f *ForgeYaml) {
	g.forge = f
}

func (r *GroupStruct) SetHandler(source string, from func(field string) (string, bool), keys ...string) {
	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(source, key, v)
		}
	}
}

func (g *GroupStruct) Set(source, field, value string) {
	switch field {
	case "role":
		if value != g.Role {
			g.Role = value
			g.forge.dirty()
		}
	case "members":
		return
	default:
		if g.More == nil {
			g.More = make(map[string]string)
		}
		if v, found := g.More[field]; found && value == "" {
			delete(g.More, field)
			g.forge.dirty()
		} else {
			if v != value {
				g.forge.dirty()
				g.More[field] = value
			}
		}
	}
	g.sources = g.sources.Set(source, field, value)
	return
}

func (g *GroupStruct) mergeFrom(from *GroupStruct) {
	for _, flag := range from.Flags() {
		if v, found, source := from.Get(flag); found {
			g.Set(source, flag, v.GetString())
		}
	}
}
