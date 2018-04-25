package scandrivers

import (
	"fmt"
	"forjj/drivers"
	"forjj/forjfile"

	"github.com/forj-oss/goforjj"
)

type ScanDrivers struct {
	ffd                *forjfile.DeployForgeYaml
	deploy             string
	missing            bool
	drivers            map[string]*drivers.Driver
	taskFlag           func(_ *forjfile.DeployForgeYaml, name string, flag goforjj.YamlFlag, missing bool) error
	objectGetInstances func(object_name string) (ret []string)
	objectFlag         func(_ *forjfile.DeployForgeYaml, object_name, instance_name, flag_prefix, name string, flag goforjj.YamlFlag, missing bool) error
}

// NewScanDrivers creates a ScanDrivers object to scan Forjfile, creds or anything through drivers flags (tasks or objects)
func NewScanDrivers(ffd *forjfile.DeployForgeYaml, drivers map[string]*drivers.Driver) (ret *ScanDrivers) {
	if ffd == nil {
		return nil
	}
	ret = new(ScanDrivers)
	ret.ffd = ffd
	ret.drivers = drivers
	return
}

// SetScanTaskFlagsFunc regsiter the taskFlag function to the scanDrivers
func (s *ScanDrivers) SetScanTaskFlagsFunc(taskFlag func(_ *forjfile.DeployForgeYaml, name string, flag goforjj.YamlFlag, missing bool) error) {
	if s == nil {
		return
	}
	s.taskFlag = taskFlag
}

// SetScanGetObjInstFunc define the objectGetInstances function used to loop on instances
func (s *ScanDrivers) SetScanGetObjInstFunc(objectGetInstances func(object_name string) (ret []string)) {
	if s == nil {
		return
	}
	s.objectGetInstances = objectGetInstances
}

// SetScanObjFlag define the objectFlag function used to loop on instances/flags or group flags
func (s *ScanDrivers) SetScanObjFlag(objectFlag func(_ *forjfile.DeployForgeYaml, object_name, instance_name, flag_prefix, name string, flag goforjj.YamlFlag, missing bool) error) {
	if s == nil {
		return
	}
	s.objectFlag = objectFlag
}

func (s *ScanDrivers) checkScanParameters() error {
	if s == nil {
		return fmt.Errorf("scanDrivers is nil")
	}
	if s.taskFlag == nil {
		return fmt.Errorf("Missing scanDrivers func 'taskFlag'")
	}
	if s.objectFlag == nil {
		return fmt.Errorf("Missing scanDrivers func 'objFlag'")
	}
	return nil
}

// DoScanDriversObject start the loop on drivers tasks and objects.
func (s *ScanDrivers) DoScanDriversObject(deploy string, missing bool) (err error) {
	if err := s.checkScanParameters(); err != nil { // No Forjfile loaded.
		return nil
	}
	s.deploy = deploy
	for _, driver := range s.drivers {
		// Tasks flags
		for _, task := range driver.Plugin.Yaml.Tasks {
			for flagName, flag := range task {
				if err := s.taskFlag(s.ffd, flagName, flag, missing); err != nil {
					return err
				}
			}
		}

		for objectName, obj := range driver.Plugin.Yaml.Objects {
			// Instances of Forj objects
			var instances []string

			if s.SetScanGetObjInstFunc == nil {
				instances = s.ffd.GetInstances(objectName)
			} else {
				instances = s.objectGetInstances(objectName)
			}

			for _, instanceName := range instances {
				// Do not set app object values for a driver of a different application.
				if objectName == "app" && instanceName != driver.InstanceName {
					continue
				}

				// Object flags
				if err := s.DispatchObjectFlags(objectName, instanceName, "", missing, obj.Flags); err != nil {
					return err
				}
				// Object group flags
				for groupName, group := range obj.Groups {
					if err := s.DispatchObjectFlags(objectName, instanceName, groupName+"-", missing, group.Flags); err != nil {
						return err
					}
				}

			}
		}
	}

	return
}

// DispatchObjectFlags is dispatching Forjfile template data between Forjfile and creds
// All plugin defined flags set with secret ON, are moving to creds
// All plugin undefined flags named with "secret_" as prefix are considered as required to be moved to
// creds
//
// The secret transfered flag value is moved to creds functions
// while in Forjfile the moved value is set to {{ .creds.<flag_name> }}
// a golang template is then used for Forfile to get the data from the default credential structure.
func (s *ScanDrivers) DispatchObjectFlags(object_name, instance_name, flag_prefix string, missing bool, flags map[string]goforjj.YamlFlag) (err error) {
	for flag_name, flag := range flags {
		if err = s.objectFlag(s.ffd, object_name, instance_name, flag_prefix+flag_name, flag_name, flag, missing); err != nil {
			return err
		}
	}
	return
}
