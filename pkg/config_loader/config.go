package config_loader

type Config interface {
	Validate() error
}

type configEntry struct {
	name string
	ptr  Config
}

type configLoader interface {
	Name() string
	Ptr() any
}

func (c configEntry) Name() string { return c.name }
func (c configEntry) Ptr() any     { return c.ptr }

func NewConfig(name string, ptr Config) configLoader {
	return configEntry{name: name, ptr: ptr}
}
