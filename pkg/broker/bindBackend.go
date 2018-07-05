package broker

import (
	"fmt"
)

type BindCallback struct {
	key string
}

func (this BindCallback) ObjectCreated(key string, obj interface{}) {
	this.key = key
	fmt.Printf("create: %s\n", key)
}

func (this BindCallback) ObjectDeleted(key string, obj interface{}) {
	this.key = key
	fmt.Printf("delete: %s\n", key)
}

func (this BindCallback) ObjectUpdated(key string, obj interface{}) {
	this.key = key
	fmt.Printf("update: %s\n", key)

}
