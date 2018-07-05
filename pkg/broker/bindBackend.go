package broker

type BindCallback struct {
	CreatedFunc func(obj interface{})
	DeletedFunc func(obj interface{})
	UpdatedFunc func(obj interface{})
}

func (this BindCallback) ObjectCreated(key string, obj interface{}) {
	this.CreatedFunc(obj)
}

func (this BindCallback) ObjectDeleted(key string, obj interface{}) {
	this.DeletedFunc(obj)
}

func (this BindCallback) ObjectUpdated(key string, obj interface{}) {
	this.UpdatedFunc(obj)
}
