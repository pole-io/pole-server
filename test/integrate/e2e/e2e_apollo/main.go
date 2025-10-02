package main

import (
	"fmt"
	"log"

	"github.com/apolloconfig/agollo/v4"
	apollolog "github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
)

// Apollo 配置模型到 Pole 配置模型的映射关系
// 1. Apollo 的 Cluster or DataCenter 映射到 Pole 的 Namespace
// 2. Apollo 的 AppId 映射到 Pole 的 Group
// 3. Apollo 的 Namespace 映射到 Pole 的 ConfigFile
func main() {
	apollolog.InitLogger(&stdLog{})

	c := &config.AppConfig{
		AppID:          "testApplication_yang",
		Cluster:        "dev",
		IP:             "http://127.0.0.1:8890",
		NamespaceName:  "dubbo.properties",
		IsBackupConfig: true,
		MustStart:      true,
	}

	client, _ := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	fmt.Println("Apollo configuration initialized successfully")

	//Use your apollo key to test
	cache := client.GetConfigCache(c.NamespaceName)
	cache.Range(func(key, value interface{}) bool {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
		return true // continue iteration
	})
	value, _ := cache.Get("key")
	fmt.Println(value)
}

type stdLog struct {
}

func (l *stdLog) Debugf(format string, params ...interface{}) {
	log.Printf("[DEBUG] "+format, params...)
}

func (l *stdLog) Infof(format string, params ...interface{}) {
	log.Printf("[INFO] "+format, params...)
}

func (l *stdLog) Warnf(format string, params ...interface{}) {
	log.Printf("[WARN] "+format, params...)
}

func (l *stdLog) Errorf(format string, params ...interface{}) {
	log.Printf("[ERROR] "+format, params...)
}

func (l *stdLog) Debug(v ...interface{}) {
	params := append([]interface{}{"[DEBUG]"}, v...)
	log.Println(params...)
}

func (l *stdLog) Info(v ...interface{}) {
	params := append([]interface{}{"[INFO]"}, v...)
	log.Println(params...)
}

func (l *stdLog) Warn(v ...interface{}) {
	params := append([]interface{}{"[WARN]"}, v...)
	log.Println(params...)
}

func (l *stdLog) Error(v ...interface{}) {
	params := append([]interface{}{"[ERROR]"}, v...)
	log.Println(params...)
}
