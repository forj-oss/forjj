package drivers

// Drivers is a collection of drivers
type Drivers struct {
	drivers map[string]*Driver
}

// Get return a loaded driver.
func (d *Drivers) Get(name string) (ret *Driver, found bool) {
	if d == nil {
		return
	}
	ret, found = d.drivers[name]
	return
}

// List return the drivers map list.
func (d *Drivers) List() (ret map[string]*Driver) {
	if d == nil {
		return
	}
	ret = d.drivers
	return
}

// Add a new Driver to the list of drivers.
func (d *Drivers) Add(name string, driver *Driver) (ret map[string]*Driver) {
	if d == nil {
		return
	}

	if d.drivers == nil {
		d.drivers = make(map[string]*Driver)
	}

	d.drivers[name] = driver
	return
}
