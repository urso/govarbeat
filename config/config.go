// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Govarbeat GovarbeatConfig
}

type GovarbeatConfig struct {
	Remotes map[string]struct {
		Period  string   `yaml:"period"`
		Hosts   []string `yaml:"host"`
		Timeout string   `yaml:"timeout"`
	}
}
