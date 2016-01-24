// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type GovarbeatConfig struct {
	Remotes map[string]struct {
		Host   string
		Period string
	}
}
