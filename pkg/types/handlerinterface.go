package types

type MsgType int

const (
	MsgType_Create MsgType = iota
	MsgType_Update MsgType = iota
	MsgType_Delete MsgType = iota
)

type CallbackHandler interface {
	ObjectCreated(key string, obj interface{})
	ObjectDeleted(key string, obj interface{})
	ObjectUpdated(key string, obj interface{})
}
