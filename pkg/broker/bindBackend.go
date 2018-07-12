package broker

import (
	"encoding/json"
	"fmt"
)

type BindCallback struct {
	key        string
	bindingMap map[string]*mySQLServiceBinding
}

func (this BindCallback) ObjectCreated(key string, obj interface{}) {
	this.key = key
	fmt.Printf("create: %s\n", key)
	var mb mySQLServiceBinding
	err := json.Unmarshal(obj.([]byte), &mb)
	if err != nil {
		fmt.Printf("damn")
	}
}

func (this BindCallback) ObjectDeleted(key string, obj interface{}) {
	this.key = key
	fmt.Printf("delete: %s\n", key)
	delete(this.bindingMap, key)
}

func (this BindCallback) ObjectUpdated(key string, obj interface{}) {
	this.key = key
	fmt.Printf("create: %s\n", key)
	var mb mySQLServiceBinding
	err := json.Unmarshal(obj.([]byte), &mb)
	if err != nil {
		fmt.Printf("damn")
	}

}

func (this BindCallback) ParseObject(data []byte) (interface{}, error) {
	var mb mySQLServiceBinding
	err := json.Unmarshal(data, &mb)
	if err != nil {
		return nil, err
	}
	return mb, err
}
