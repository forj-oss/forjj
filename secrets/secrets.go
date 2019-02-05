package secrets

import (
	"fmt"
	"forjj/creds"
	"forjj/drivers"
	"forjj/forjfile"
	"forjj/utils"
	"io/ioutil"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/cli/clier"
	"github.com/forj-oss/goforjj"
)

type Secrets struct {
	secrets *kingpin.CmdClause
	Context Context
	common  common

	list sList

	get struct {
		cmd *kingpin.CmdClause
		key *string
	}

	set sSet

	edit sEdit

	unset sUnset
}

// Init initialize the secrets cli commands
func (s *Secrets) Init(app *kingpin.Application, forjfile *forjfile.Forge, drivers *drivers.Drivers, secrets *creds.Secure, isParsePhase func() bool, initCommon func(context *Context, cmd *kingpin.CmdClause)) {
	if s == nil || app == nil {
		return
	}

	s.secrets = app.Command("secrets", "Manage forjj secrets")
	s.Context.init(isParsePhase)
	s.common.init(&s.Context, s.secrets, initCommon)
	s.list.init(s.secrets, &s.common, forjfile, drivers, secrets)
	s.set.init(s.secrets, &s.common, forjfile, drivers, secrets)
	s.edit.init(s.secrets, &s.common, forjfile, drivers, secrets)

	s.get.cmd = s.secrets.Command("get", "Get value of a credential unencrypted")
	s.get.key = s.get.cmd.Arg("key", "Full key path").Required().String()

	s.unset.init(s.secrets, &s.common, forjfile, drivers, secrets)
}

func (s *Secrets) Action(action string) {
	actions := strings.Split(action, " ")
	switch actions[1] {
	case "list":
		s.list.showList()
	case "set":
		s.set.doSet()
	case "edit":
		s.edit.doEdit()
	case "unset":
		s.unset.doUnset()
	case "show":
	}
}

// DefineContext define cli Context to permit ParseContext to retrieve
// common variable set.
func (s *Secrets) DefineContext(context clier.ParseContexter) {
	s.Context.defineContext(context)
}

// GetStringValue Return a field value from the given context (parse time, or after)
func (s *Secrets) GetStringValue(field string) (value string, found, isDefault bool, _ error) {
	return s.Context.GetStringValue(field)
}

// DefineSetters define the list of secrets value setters (ex: copied-from, link-to)
func DefineSetters(s *creds.Secure) {

	// Define how secret set a file link
	s.SetSetterHandler(link, func(v *creds.Value, value *goforjj.ValueStruct) (_ error) {
		file := value.GetString()

		if f, err := utils.Abs(file); err != nil {
			return fmt.Errorf("Unable to determine file path. %s", err)
		} else {
			file = f
		}

		v.SetValue(goforjj.NewValueStruct(""))
		v.AddResource("linked-to", file)
		v.AddResource("created-on", time.Now().String())

		return
	})

	// Define how secret set a file copy
	s.SetSetterHandler(copy, func(v *creds.Value, value *goforjj.ValueStruct) (_ error) {
		file := value.GetString()

		if f, err := utils.Abs(file); err != nil {
			return fmt.Errorf("Unable to determine file path. %s", err)
		} else {
			file = f
		}

		data, err := ioutil.ReadFile(file)
		if err != nil {
			err = fmt.Errorf("Unable to copy file content to forjj secret. %s", err)
			return
		}

		v.SetValue(goforjj.NewValueStruct(string(data)))
		v.AddResource("copied-from", file)
		v.AddResource("copied-on", time.Now().String())

		return
	})

}

// DefineGetters define the list of secrets value setters (ex: copied-from, link-to)
func DefineGetters(s *creds.Secure) {

	// Define how secrets get a link
	s.SetGetterHandler(link, func(v *creds.YamlValue) (value string, err error) {
		file, found := v.Resource["linked-to"]
		if !found {
			err = fmt.Errorf("Invalid Secret type. 'linked-to' resource is empty")
			return
		}

		data, err := ioutil.ReadFile(file)
		if err != nil {
			err = fmt.Errorf("Unable to copy file content to forjj secret. %s", err)
			return
		}

		value = string(data)

		return
	})

	// Define how secrets get a copy
	s.SetGetterHandler(copy, func(v *creds.YamlValue) (value string, _ error) {
		value = v.Value.GetString()

		return
	})

}
