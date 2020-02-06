package config

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type App struct {
	cfg  *Config
	lock sync.Mutex
}

func New(cfg *Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Config() *Config {
	a.lock.Lock()
	cfg := a.cfg
	a.lock.Unlock()
	return cfg
}

func (a *App) SetConfig(cfg *Config) {
	a.lock.Lock()
	a.cfg = cfg
	a.lock.Unlock()
}

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
//		x := App{}
//		x.setConfig(<-cfgCh)
//
//		fmt.Println(&c)
//
//		time.Sleep(2 * time.Second)
//
//	}
//
//}
