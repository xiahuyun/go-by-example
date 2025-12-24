package main

import (
	"fmt"
	"sync"
)

type Config struct {
	config map[string]string
}

type CowConfig struct {
	mtx    sync.Mutex
	config Config
}

func NewCowConfig() *CowConfig {
	return &CowConfig{
		mtx: sync.Mutex{},
		config: Config{
			config: make(map[string]string),
		},
	}
}

func (c *Config) Get() map[string]string {
	return c.config
}

func (c *Config) Update(key, value string) {
	c.config[key] = value
}

func (cc *CowConfig) Get() map[string]string {
	//cc.mtx.Lock()
	//defer cc.mtx.Unlock()

	return cc.config.Get()
}

func (cc *CowConfig) Update(key string, value string) {
	if v, ok := cc.Get()[key]; ok {
		if v != value {
			fmt.Println("updating")

			cc.mtx.Lock()
			cc.config.Update(key, value)
			cc.mtx.Unlock()
		}
	}

	cc.mtx.Lock()
	cc.config.Update(key, value)
	cc.mtx.Unlock()
}

func main() {
	config := NewCowConfig()

	config.Update("name", "hxia")
	fmt.Println(config.Get())

	config.Update("name", "hxia")
	fmt.Println(config.Get())
}
