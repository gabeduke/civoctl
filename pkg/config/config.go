package config

type Config struct {
	Clusters []struct {
		Name  string `yaml:"name"`
		Nodes int    `yaml:"nodes"`
	} `yaml:"clusters"`
}
