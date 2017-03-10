package forjfile

import "strings"

type GroupStruct struct {
	forge *Forge
	Role string
	Members []string
	More map[string]string `yaml:",inline"`
}

// TODO: Add struct unit tests

func (g *GroupStruct) Get(field string) (value string, found bool) {
	switch field {
	case "role":
		return g.Role, (g.Role != "")
	case "members":
		if g.Members != nil && len(g.Members) > 0 {
			return strings.Join(g.Members, ","), true
		}
		return "", false
	default:
		value, found = g.More[field]
	}
	return
}

func (g *GroupStruct) GetMembers() []string {
	return g.Members
}

func (g *GroupStruct) AddMembers(members ...string) (count int) {
	add_members := map[string]int{}
	for _, new_member := range members {
		if g.hasMember(new_member) >=0 {
			add_members[new_member]=0
		}
	}
	count = len(add_members)
	if (count > 0) {
		g.forge.dirty()
	}
	new_members := make([]string, 0, len(add_members) + len(g.Members))
	copy(new_members, g.Members)
	copy(new_members, keys(add_members))
	g.Members = new_members
	return
}

func (g *GroupStruct) RemoveMembers(members ...string) (count int) {
	for _, new_member := range members {
		if index := g.hasMember(new_member) ; index >=0 {
			g.Members = removeSliceString(g.Members, index)
			count++
		}
	}
	if (count > 0) {
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

func (g *GroupStruct) set_forge(f *Forge) {
	g.forge = f
}

func (g *GroupStruct) Set(field, value string) {
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
		if v, found := g.More[field] ; found && v != value {
			g.More[field] = value
			g.forge.dirty()
		}
	}
	return
}
