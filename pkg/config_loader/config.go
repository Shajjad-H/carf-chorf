package config_loader

type Config interface {
	Validate() error
}

type configEntry struct {
	name     string
	ptr      Config
	yamlPath string
}

type configLoader interface {
	Name() string
	Ptr() any
	YAMLPath() string
}

func (c configEntry) Name() string     { return c.name }
func (c configEntry) Ptr() any         { return c.ptr }
func (c configEntry) YAMLPath() string { return c.yamlPath }

func NewConfig(name string, ptr Config) configLoader {
	return configEntry{name: name, ptr: ptr}
}

func NewYAMLConfig(name, path string, ptr Config) configLoader {
	return configEntry{name: name, ptr: ptr, yamlPath: path}
}
