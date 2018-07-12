package broker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"reflect"

	"github.com/cohenjo/mysql-osb/pkg/types"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

// see here a nice example: https://github.com/opensds/nbp/blob/master/service-broker/controller/controller.go

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.

	b := &BusinessLogic{
		async:              o.Async,
		dbConnectionString: o.DBConnectionString,
		etcClient:          NewEtcdClient(3),
		instances:          make(map[string]*dbInstance, 10),
		bindingMap:         make(map[string]*mySQLServiceBinding),
	}
	b.initSchema()
	b.initWatchers()
	return b, nil
}

type mySQLServiceBinding struct {
	ID, InstanceID, ServiceID, PlanID string
	artifact                          string
	cluster                           string
	BindResource                      *osb.BindResource
	Params                            map[string]interface{}
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// Add fields here! These fields are provided purely as an example
	dbConnectionString string
	etcClient          *EtcdClientAPIv3
	instanceWatcher    *EtcdWatcher
	bindWatcher        *EtcdWatcher
	instances          map[string]*dbInstance
	bindingMap         map[string]*mySQLServiceBinding
}

var _ broker.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog business logic goes here
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "mysql-artifact-service",
				ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad246a",
				Description:   "Provision a MySQL cluster.",
				Bindable:      true,
				PlanUpdatable: truePtr(),
				Metadata: map[string]interface{}{
					"displayName": "Provision a MySQL cluster",
					"imageUrl":    "https://planet.mysql.com/images/planet-logo.svg",
				},
			},
		},
	}
	osbResponse.Services[0].Plans = []osb.Plan{
		types.Plan,
	}

	glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	exampleInstance := &dbInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}
	glog.V(4).Infof("InstanceL %s !\n", request.InstanceID)
	glog.V(4).Infof("service:  %s !\n", request.ServiceID)
	glog.V(4).Infof("plan:     %s !\n", request.PlanID)
	glog.V(4).Infof("Paramet:  %s !\n", request.Parameters)

	// Check to see if this is the same instance
	if i := b.instances[request.InstanceID]; i != nil {
		if i.Match(exampleInstance) {
			response.Exists = true
			return &response, nil
		} else {
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		}
	}

	b.etcIt(request, exampleInstance)
	go b.order(request, exampleInstance)
	// go

	b.instances[request.InstanceID] = exampleInstance

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	// Your deprovision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	delete(b.instances, request.InstanceID)

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	// Your last-operation business logic goes here

	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	// Your bind business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := &broker.BindResponse{}

	if _, ok := b.bindingMap[request.BindingID]; ok {
		glog.Infof("Binding %s already exist!\n", request.BindingID)
		response.Exists = true
		return response, nil
	}

	instance, ok := b.instances[request.InstanceID]
	if !ok {
		errMsg := fmt.Sprintf("Instance (%s) not found in instance map!", request.InstanceID)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusNotFound,
			ErrorMessage: &errMsg,
		}
	}

	clusterID, ok := instance.Params["cluster"]
	if !ok {
		errMsg := fmt.Sprint("hostInfo not found in bind request params!")
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}

	newBinding := &mySQLServiceBinding{
		artifact:     request.Parameters["artifact"].(string),
		cluster:      request.Parameters["cluster"].(string),
		InstanceID:   request.InstanceID,
		BindResource: request.BindResource,
	}

	b.bindingMap[request.BindingID] = newBinding
	glog.Infof("Created mysql Service Binding:\n%v\n",
		b.bindingMap[request.BindingID])
	b.etcClient.EtcIt(request.BindingID, newBinding)

	// Generate service binding credentials.
	creds := make(map[string]interface{})
	creds["user"] = "root"
	creds["password"] = "password"
	creds["db"] = "request db"
	creds["cluster"] = clusterID

	osbResponse := &osb.BindResponse{
		Credentials: creds,
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}
	response.BindResponse = *osbResponse
	return response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}

// example types

// dbInstance is a type that holds information about a db instance
type dbInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
}

func (i *dbInstance) Match(other *dbInstance) bool {
	return reflect.DeepEqual(i, other)
}

func (i *dbInstance) String() (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
