package drivers

// DriverModel Structure used as template context. The way to get it: Driver.Model()
type DriverModel struct {
	InstanceName string
	Name         string
}

// Model return the driver object model for text/template
func (d *Driver) Model() (m *DriverModel) {
	m = &DriverModel{
		InstanceName: d.InstanceName,
		Name:         d.Name,
	}
	return
}
