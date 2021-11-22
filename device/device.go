package device

// Device represents a device, with a name easy to remember and the UUID to identify it, optionally an owner.
type Device struct {
	Name      string `mapstructure:",omitempty"`
	UUID      string `mapstructure:"uuid"`
	OwnerUser string `mapstructure:"owner,omitempty"`
	isMounted bool
}

// Mount will mount a device
func (d Device) Mount() {

	d.isMounted = true
}

// Unmount will unmount a device
func (d Device) Unmount() {

	d.isMounted = false
}

func (d Device) IsMounted() bool {
	return d.isMounted
}
