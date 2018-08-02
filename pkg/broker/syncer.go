package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cohenjo/mysql-osb/pkg/types"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/glog"
	"github.com/sirupsen/logrus"

	"errors"
)

type EtcdClientAPIv3 struct {
	client       *clientv3.Client
	log          *logrus.Entry
	timeout      time.Duration
	NumOfRetries int
}

type EtcdWatcher struct {
	client   *clientv3.Client
	Folder   string
	Index    uint64
	callback types.CallbackHandler
	cancel   context.CancelFunc
	log      *logrus.Entry

	// dataStore  *types.DataStore
	objectType types.ObjectType

	lock sync.RWMutex
}

func GenerateEtcdClient(addressPool []string) (retVal *clientv3.Client, err error) {
	glog.V(4).Infof("Creating client")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   addressPool,
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		glog.V(4).Infof("error with etcd, %s !\n", err.Error())
		// panic(err.Error())
	}
	return cli, err
}

func NewEtcdClient2(addressPool []string, numOfRetries int) (retVal *EtcdClientAPIv3) {
	log := logrus.WithField("obj", "etcdv3")

	client, err := GenerateEtcdClient(addressPool)
	if err != nil {
		glog.V(4).Infof("error with etcd, %s !\n", err.Error())
		// panic(err.Error())
	}
	glog.V(4).Infof("created client,  !\n")

	retValObj := &EtcdClientAPIv3{}

	retValObj.client = client

	retValObj.log = log
	retValObj.timeout = time.Second * 20
	retValObj.NumOfRetries = numOfRetries
	retVal = retValObj
	return retVal
}

func NewEtcdClient1(adress string, numOfRetries int) (retVal *EtcdClientAPIv3) {
	addressPool := []string{adress}
	return NewEtcdClient2(addressPool, numOfRetries)
}

func NewEtcdClient(numOfRetries int) (retVal *EtcdClientAPIv3) {
	addressPool := []string{"etcd-cluster-client:2379"}
	return NewEtcdClient2(addressPool, numOfRetries)
}

func NewEtcdWatcher(clientPar *EtcdClientAPIv3, objectTypePar types.ObjectType, indexPar uint64, callbackPar types.CallbackHandler) (retVal *EtcdWatcher) {
	var folder string
	folder = GetFolderName(objectTypePar)

	retVal = NewEtcdWatcherWithKey(clientPar, folder, indexPar, callbackPar)
	retVal.objectType = objectTypePar
	return retVal
}

func NewEtcdWatcherWithKey(clientPar *EtcdClientAPIv3, keyToWatch string, indexPar uint64, callbackPar types.CallbackHandler) (retVal *EtcdWatcher) {

	client := clientPar.GetClientObj().(*clientv3.Client)

	retVal = &EtcdWatcher{
		client:   client,
		Folder:   keyToWatch,
		Index:    indexPar,
		callback: callbackPar,
		// dataStore: storePar,
		log:        logrus.WithFields(logrus.Fields{"obj": "etcdv3", "watchtype": keyToWatch}),
		objectType: types.Binding,
		lock:       sync.RWMutex{},
	}
	return retVal
}

func (this EtcdClientAPIv3) Close() {
	this.client.Close()
}

func (this EtcdClientAPIv3) GetClientObj() (retVal interface{}) {

	return this.client
}

func (this EtcdClientAPIv3) Get(keyname string) (retVal string, err error) {
	getResponse, err := this.client.Get(context.Background(), keyname)
	if err == nil && len(getResponse.Kvs) > 0 {
		retVal = string(getResponse.Kvs[0].Value)

	} else {
		err = errors.New("Key Not Found")
	}
	return retVal, err
}

/*
func (this EtcdClientAPIv3) UploadFile(keyname string, filename string) (err error) {
	this.log.Printf("Uploading file %s to key %s", filename, keyname)
	var fileContent []byte
	fileContent, err = ioutil.ReadFile(filename)
	if err != nil {
		this.log.Error(err)
		return err
	}
	err = this.Set(keyname, string(fileContent))
	return err
}
*/

func (this EtcdClientAPIv3) EtcIt(keyname string, value interface{}) (err error) {
	b, err := json.Marshal(value)
	if err != nil {
		fmt.Println("error:", err)
	}
	newkey := fmt.Sprintf("/db/mysql-broker/%s", keyname)
	glog.V(4).Infof("Creating key: %s in etcd", newkey)
	_, err = this.client.Put(context.Background(), newkey, string(b))
	return err
}

func (this EtcdClientAPIv3) Set(keyname string, value string) (err error) {
	_, err = this.client.Put(context.Background(), keyname, value)
	return err
}

// func (this EtcdClientAPIv3) SetServiceImage(image types.ServiceImage) (err error) {
// 	serviceID := image.GetServiceID()
// 	localLog := this.log.WithFields(logrus.Fields{"action": "setserviceimage", logging.FieldName_ServiceId: serviceID})
// 	localLog.Infof("Setting new value for service %s", serviceID)
// 	attemptsCount := 0
// 	for attemptsCount < this.NumOfRetries {
// 		err = this.Set(path.Join(GetFolderName(types.Image), serviceID), image.GetRawData())
// 		if err == nil {
// 			break
// 		}
// 		attemptsCount++
// 	}
// 	if err != nil {
// 		localLog.WithField("result", "failure").Errorf("%s Attempt %d", err, attemptsCount)
// 	} else {
// 		localLog.WithField("result", "ok").Infof("Record updated for service %s", serviceID)
// 	}

// 	return err
// }

// func (this EtcdClientAPIv3) GetListOfObjects(objectType types.ObjectType) (retVal []interface{}, lastIndex uint64, err error) {
// 	var parentFolder string
// 	parentFolder = GetFolderName(objectType)

// 	opts := []clientv3.OpOption{clientv3.WithPrefix()}

// 	resp, err := this.client.Get(context.Background(), parentFolder, opts...)

// 	for _, ev := range resp.Kvs {
// 		if objectType == types.Image {
// 			retValCurrent, err := types.NewServiceImageFromEtcdRawData(ev.Value)
// 			if err == nil {
// 				retVal = append(retVal, retValCurrent)
// 			} else {
// 				this.log.Error("Error initializing ServiceImage ", err)
// 			}
// 		} else if objectType == types.Definition {
// 			retValCurrent, err := types.NewServiceDefinitionFromRawData(ev.Value)
// 			if err == nil {

// 				retVal = append(retVal, retValCurrent)
// 			} else {
// 				this.log.Error("Error initializing ServiceDefinition", err)
// 			}

// 		}
// 	}
// 	return retVal, 0, err
// }

func GetFolderName(objectTypePar types.ObjectType) (retVal string) {
	retVal = fmt.Sprintf("/db/mysql-broker/%s/", string(objectTypePar))
	return retVal
}

func (this *EtcdWatcher) RunAsync() (err error) {

	contextWithCancelm, cancel := context.WithCancel(context.Background())
	// var requestTimeout = 10 * time.Second
	// contextWithCancelm, _ := context.WithTimeout(context.Background(), requestTimeout)
	this.cancel = cancel
	go func() {
		for {
			this.log.Infof("Watch started on %s", this.objectType)
			watchChan := this.client.Watch(contextWithCancelm, GetFolderName(this.objectType), clientv3.WithPrefix())
			r, open := <-watchChan
			if !open {
				this.log.Infof("Watch channel closed. Wait before retry watching...")
			} else {
				receivedEvents := r.Events
				this.log.Infof("Events received: %d", len(receivedEvents))
				for _, event := range receivedEvents {
					handleWatcherResponseEvent(event, this)
				}
				this.log.Info("Done")
			}

			select {
			case <-contextWithCancelm.Done():
				this.log.Info("watcher canceled")
				return
			case <-time.After(5 * time.Second):
				this.log.Info("Retry watch")
			}

		}

	}()
	return err
}

func (this EtcdWatcher) ReloadCacheData() (lastIndex uint64, err error) {
	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	resp, err := this.client.Get(context.Background(), GetFolderName(this.objectType), opts...)

	for _, ev := range resp.Kvs {
		// var retVal interface{}
		// serviceID := ""
		if this.objectType == types.Binding {
			this.log.Infof("reload binding: %s", ev.Key)
		} else if this.objectType == types.Instance {
			this.log.Infof("reload instance: %s", ev.Key)
			this.callback.ObjectCreated(string(ev.Key), ev.Value)
		}
	}

	return 0, err
}

func (watcher EtcdWatcher) CancelWait() {
	watcher.log.Debug("Cancel wait")
	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	if watcher.cancel != nil {
		watcher.cancel()
	}
}

func handleWatcherResponseEvent(event *clientv3.Event, this *EtcdWatcher) {

	var valueToParse []byte
	var callbackToExecute func(key string, obj interface{})
	var actioinTypeMarker string
	if event.Type == 1 {
		valueToParse = event.PrevKv.Value
		callbackToExecute = this.callback.ObjectDeleted
		actioinTypeMarker = "D"
	} else {
		valueToParse = event.Kv.Value
		if event.Kv.Version == 1 {
			callbackToExecute = this.callback.ObjectCreated
			actioinTypeMarker = "C"
		} else {
			callbackToExecute = this.callback.ObjectUpdated
			actioinTypeMarker = "U"
		}
	}
	callbackToExecute(string(event.Kv.Key), valueToParse)
	this.log.Infof("handled event: %s, for key: %s\n", actioinTypeMarker, string(event.Kv.Key))
}
