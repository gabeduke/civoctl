package civo

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"sync"
	"time"
)

// CivoCtl holds the config and interface for a Civo Controller
type CivoCtl struct {
	Client    *Civo
	cfg       *Config
	lock      sync.Mutex
	Dangerous bool
}

// Config contains the list of clusters CivoCtl will handle
type Config struct {
	Clusters []struct {
		Name  string `yaml:"name"`
		Nodes int    `yaml:"nodes"`
	} `yaml:"clusters"`
}

// NewCivoCtl configures a Civo interface
func NewCivoCtl(cfg *Config, token string, dangerous bool) *CivoCtl {
	civo := newCivoHandler(token)
	return &CivoCtl{
		Client:    civo,
		cfg:       cfg,
		Dangerous: dangerous,
	}
}

// Config returns a safe copy of CivoCtl cfg
func (a *CivoCtl) Config() *Config {
	a.lock.Lock()
	cfg := a.cfg
	a.lock.Unlock()
	return cfg
}

// SetConfig sets a safe copy of CivoCtl cfg
func (a *CivoCtl) SetConfig(cfg *Config) {
	a.lock.Lock()
	a.cfg = cfg
	a.lock.Unlock()
}

// LoadConfig returns a config from viper and updates channel
func LoadConfig() (*Config, chan *Config) {

	//TODO check for nil cfg

	config := &Config{}
	viper.Unmarshal(config)

	viper.WatchConfig()

	configCh := make(chan *Config, 1)

	prev := time.Now()
	viper.OnConfigChange(func(e fsnotify.Event) {
		now := time.Now()
		// fsnotify sometimes fires twice
		if now.Sub(prev) > time.Second {
			config := &Config{}
			err := viper.Unmarshal(config)
			if err == nil {
				configCh <- config
			}

			prev = now
		}
	})

	return config, configCh
}

//func usage() {
//
//	for {
//
//		c, cfgCh := LoadConfig()
//
//		x := CivoCtl{}
//		x.setConfig(<-cfgCh)
//
//		fmt.Println(&c)
//
//		time.Sleep(2 * time.Second)
//
//	}
//
//}
