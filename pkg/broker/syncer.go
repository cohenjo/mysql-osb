package broker

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/glog"

	"errors"
)

type EtcdClientAPIv3 struct {
	client       *clientv3.Client
	timeout      time.Duration
	NumOfRetries int
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

func NewEtcdClient(numOfRetries int) (retVal *EtcdClientAPIv3) {
	// log := logging.GetLog("etcd")
	addressPool := []string{"etcd-cluster-client:2379"}
	client, err := GenerateEtcdClient(addressPool)
	if err != nil {
		glog.V(4).Infof("error with etcd, %s !\n", err.Error())
		// panic(err.Error())
	}
	glog.V(4).Infof("created client,  !\n")

	retValObj := &EtcdClientAPIv3{}

	retValObj.client = client

	// retValObj.log = log
	retValObj.timeout = time.Second * 20
	retValObj.NumOfRetries = numOfRetries
	retVal = retValObj
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
