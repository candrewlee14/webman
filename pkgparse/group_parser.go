package pkgparse

type PkgGroupConfig struct {
	Title   string
	Tagline string
	About   string

	InfoUrl  string   `yaml:"info_url"`
	Packages []string `yaml:"packages"`
}
