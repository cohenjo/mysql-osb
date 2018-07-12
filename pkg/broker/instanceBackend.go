package broker

import (
	"encoding/json"
	"fmt"
)

type InstanceCallback struct {
	key            string
	InstanceingMap map[string]*dbInstance
}

func (this InstanceCallback) ObjectCreated(key string, obj interface{}) {
	this.key = key
	fmt.Printf("create: %s\n", key)
	var mb dbInstance
	err := json.Unmarshal(obj.([]byte), &mb)
	if err != nil {
		fmt.Printf("damn")
	}
	this.InstanceingMap[key] = &mb
}

func (this InstanceCallback) ObjectDeleted(key string, obj interface{}) {
	this.key = key
	fmt.Printf("delete: %s\n", key)
	delete(this.InstanceingMap, key)
}

func (this InstanceCallback) ObjectUpdated(key string, obj interface{}) {
	this.key = key
	fmt.Printf("create: %s\n", key)
	var mb dbInstance
	err := json.Unmarshal(obj.([]byte), &mb)
	if err != nil {
		fmt.Printf("damn")
	}
	this.InstanceingMap[key] = &mb

}

func (this InstanceCallback) ParseObject(data []byte) (interface{}, error) {
	var mb dbInstance
	err := json.Unmarshal(data, &mb)
	if err != nil {
		return nil, err
	}
	return mb, err
}
