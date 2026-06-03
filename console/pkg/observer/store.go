package store

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// Config Store的通用配置
type Config struct {
	Name   string
	Option map[string]interface{}
}

var (
	// StoreSlots store slots
	StoreSlots = make(map[string]ObserverStore)

	once   = &sync.Once{}
	config = &Config{}
)

// RegisterStore 注册一个新的Store
func RegisterStore(s ObserverStore) error {
	name := s.Name()
	if _, ok := StoreSlots[name]; ok {
		return errors.New("store name already existed")
	}

	StoreSlots[name] = s
	return nil
}

// GetStore 获取Store
func GetStore() (ObserverStore, error) {
	name := config.Name
	if name == "" {
		return nil, errors.New("store name is empty")
	}

	store, ok := StoreSlots[name]
	if !ok {
		return nil, fmt.Errorf("store `%s` not found", name)
	}

	initialize(store)
	return store, nil
}

// SetStoreConfig 设置store的conf
func SetStoreConfig(conf *Config) {
	config = conf
}

// initialize  包裹了初始化函数，在GetStore的时候会在自动调用，全局初始化一次
func initialize(s ObserverStore) {
	once.Do(func() {
		fmt.Printf("[Store][Info] current use store plugin : %s\n", s.Name())
		if err := s.Initialize(config); err != nil {
			fmt.Printf("[ERROR] initialize store `%s` fail: %v", s.Name(), err)
			os.Exit(1)
		}
	})
}
