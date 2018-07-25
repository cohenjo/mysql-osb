package broker

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"strconv"

	"github.com/cohenjo/mysql-osb/pkg/types"
	_ "github.com/go-sql-driver/mysql" // import the mysql driver
	"github.com/golang/glog"

	"github.com/google/uuid"
	"k8s.io/api/apps/v1beta1"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mitchellh/mapstructure"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	clientset "k8s.io/client-go/kubernetes"
	clientrest "k8s.io/client-go/rest"
)

type Order struct {
	InstanceID string
	ServiceID  string
	PlanID     string
}

type Parameters struct {
	Artifact       string
	DeploymentType string
	Size           int
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (b *BusinessLogic) initSchema() {
	db, err := sql.Open("mysql", b.dbConnectionString)
	if err != nil {
		glog.V(4).Infof("error with db !\n")
		panic(err.Error())
	}
	defer db.Close()

	t := `CREATE TABLE IF NOT EXISTS broker.orders (InstanceID VARCHAR(64) NOT NULL,
serviceID  VARCHAR(64) NOT NULL, 
PlanID VARCHAR(64) NOT NULL,
Artifact varchar(256),
DeploymentType varchar(256),
Size integer
);`
	_, err = db.Exec(t)
	if err != nil {
		glog.V(4).Infof("error with db !\n")
		panic(err.Error())
	}
}

// initWatchers spins up watchers for Bind & instance information over etcd.
func (b *BusinessLogic) initWatchers() {
	glog.V(4).Infof("Init Bind watcher")
	bindcallback := &BindCallback{}
	bindwatcher := NewEtcdWatcher(b.etcClient, types.Binding, 0, bindcallback)
	bindwatcher.ReloadCacheData()
	bindwatcher.RunAsync()
	b.bindWatcher = bindwatcher

	glog.V(4).Infof("Init instance watcher")
	callback := &InstanceCallback{
		InstanceingMap: make(map[string]*dbInstance, 10),
		bl:             b,
	}
	watcher := NewEtcdWatcher(b.etcClient, types.Instance, 0, callback)
	watcher.ReloadCacheData()
	watcher.RunAsync()
	b.instanceWatcher = watcher
	glog.V(4).Infof("All watchers Initialized")
}

func (b *BusinessLogic) etcIt(request *osb.ProvisionRequest, i *dbInstance) {
	glog.V(4).Infof("storing InstanceL in etcd: %s !\n", request.InstanceID)
	key := fmt.Sprintf("/db/mysql-broker/instance/%s", i.Params["cluster"].(string))
	value, _ := i.String()

	glog.V(4).Infof("######################################################################################################")
	glog.V(4).Infof("######################################################################################################")
	glog.V(4).Infof("put in etcd:  %s, %s !\n", key, value)
	glog.V(4).Infof("####################################### Created client ###################################")
	err := b.etcClient.Set(key, value)
	if err != nil {
		glog.V(4).Infof("error with etcd !\n")
		// panic(err.Error())
	}

	glog.V(4).Infof("Done")

	gfe, err := b.etcClient.Get(key)
	if err != nil {
		glog.V(4).Infof("error with etcd !\n")
		// panic(err.Error())
	}
	glog.V(4).Infof("got: %s \n", gfe)
	glog.V(4).Infof("Done")
	glog.V(4).Infof("######################################################################################################")
	glog.V(4).Infof("######################################################################################################")
}

func (b *BusinessLogic) order(request *osb.ProvisionRequest, i *dbInstance) {

	req := fmt.Sprintf("%v", request)
	glog.Infof("Got instance request: %s!\n", req)
	glog.V(4).Infof("service:  %s !\n", request.ServiceID)
	glog.V(4).Infof("plan:     %s !\n", request.PlanID)
	glog.V(4).Infof("Paramet:  %s !\n", request.Parameters)
	db, err := sql.Open("mysql", b.dbConnectionString)
	if err != nil {
		glog.V(4).Infof("error with db !\n")
		panic(err.Error())
	}
	defer db.Close()

	var p Parameters
	err = mapstructure.Decode(request.Parameters, &p)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	glog.V(4).Infof("Debug1")
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	glog.V(4).Infof("Debug: pinged")

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO orders VALUES( ?, ? , ? ,?, ?, ?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates
	glog.V(4).Infof("Debug2")

	_, err = stmtIns.Exec(request.InstanceID, request.ServiceID, request.PlanID, p.Artifact, p.DeploymentType, p.Size) // Insert tuples
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	glog.V(4).Infof("Debug: Pre-Select")
	results, err := db.Query("SELECT InstanceID, ServiceID, PlanID  FROM orders")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	glog.V(4).Infof("Debug: cursoe")
	for results.Next() {
		var tag Order
		// for each row, scan the result into our tag composite object
		err = results.Scan(&tag.InstanceID, &tag.ServiceID, &tag.PlanID)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		// and then print out the tag's Name attribute
		glog.V(4).Infof("select: %s !\n", tag.InstanceID)
	}

	k8sClient, err := getKubernetesClient("")
	if err != nil {
		glog.V(4).Infof("can't create a client - PANIC")
		panic(err.Error())
	}

	ret := i.GenerateService()

	svc, err := k8sClient.CoreV1().Services(b.dbNamespace).Create(&ret)
	if err != nil {
		glog.V(4).Infof("can't create a service - PANIC")
		// panic(err.Error())
		glog.V(4).Infof("can't create a service - Don't panic it exists")
	}
	glog.V(4).Infof("Debug: service status %s\n", svc.Status.String())

	cfm := i.GenerateMySQLConfigMap()
	_, err = k8sClient.CoreV1().ConfigMaps(b.dbNamespace).Create(&cfm)
	if err != nil {
		glog.V(4).Infof("can't create a config map - PANIC, %s\n", err.Error())
		// fmt.Println("fuck")
		// panic(err.Error())
		glog.V(4).Infof("can't create a service - Don't panic it exists")
	}

	retss := i.GenerateStatefulSets()

	_, err = k8sClient.AppsV1beta1().StatefulSets(b.dbNamespace).Create(&retss)
	if err != nil {
		glog.V(4).Infof("can't create a StatefulSets - PANIC")
		glog.V(4).Infof(err.Error())
		fmt.Println("fuck")
		// panic(err.Error())
	}

	glog.V(4).Infof("Debug: Done")

}

/*
GenerateStatefulSets generates something
*/
func (i *dbInstance) GenerateStatefulSets() (retVal v1beta1.StatefulSet) {
	var fileContent []byte
	parsedData := v1beta1.StatefulSet{}

	fileContent, err := ioutil.ReadFile(path.Join("/opt/servicebroker/templates", "mysql.json"))
	if err != nil {
		print(err)
		glog.V(4).Infof("Failed to read file! - Panic")
		panic(err.Error())
	}
	err = json.Unmarshal(fileContent, &parsedData)

	if err != nil {
		print(err)
	}
	fmt.Println(parsedData.GetName())
	parsedData.SetName("mysql-" + i.Params["cluster"].(string))
	var labels map[string]string
	labels = make(map[string]string)
	labels["app"] = "mysql-" + i.Params["cluster"].(string)
	parsedData.SetLabels(labels)
	fmt.Println(parsedData.GetName())

	for _, vol := range parsedData.Spec.Template.Spec.Volumes {
		if vol.Name == "config-map" {
			vol.ConfigMap.Name = "mysql-" + i.Params["cluster"].(string)
		}
	}
	parsedData.Spec.Selector.MatchLabels = labels
	parsedData.Spec.Template.ObjectMeta.Labels = labels
	glog.V(4).Infof("######################################################################################################")
	// glog.V(4).Infof("######################################################################################################")
	// glog.V(4).Infof(parsedData.String())
	// glog.V(4).Infof("######################################################################################################")
	glog.V(4).Infof("######################################################################################################")
	return parsedData
}

/**
GenerateMySQLConfigMap generates the configuration map for a new cluster
*/
func (i *dbInstance) GenerateMySQLConfigMap() (retVal api_v1.ConfigMap) {
	var fileContent []byte
	parsedData := api_v1.ConfigMap{}
	numServers := int(i.Params["NumOfReplicas"].(float64))

	fileContent, err := ioutil.ReadFile(path.Join("/opt/servicebroker/templates", "config.json"))
	if err != nil {
		glog.V(4).Infof("Failed to read config map file! - Panic")
		print(err)
	}
	err = json.Unmarshal(fileContent, &parsedData)

	if err != nil {
		glog.V(4).Infof("Failed to UnMarshal! - Panic")
		print(err)
	}
	parsedData.SetName("mysql-" + i.Params["cluster"].(string))
	var labels map[string]string
	labels = make(map[string]string)
	labels["app"] = "mysql-" + i.Params["cluster"].(string)
	parsedData.SetLabels(labels)

	t := template.Must(template.ParseFiles(path.Join("/opt/servicebroker/templates", "./my.cnf.tmpl")))
	guid := uuid.New()
	for index := 0; index < numServers; index++ {

		data := struct {
			ServerID  int
			LocalIP   string
			SeedIP    string
			GroupUUID string
		}{
			ServerID:  index,
			LocalIP:   "mysql-" + i.Params["cluster"].(string) + "-" + strconv.Itoa(index),
			SeedIP:    "mysql-" + i.Params["cluster"].(string) + "-0",
			GroupUUID: guid.String(),
		}

		var tpl bytes.Buffer
		err := t.ExecuteTemplate(&tpl, "config", data)
		if err != nil {
			glog.V(4).Infof("Failed to execute template ! - Panic")
			print(err)
		}

		result := tpl.String()
		parsedData.Data["mysql-"+strconv.Itoa(index)] = result
	}

	return parsedData
}

/*
GenerateHelloService generates something
*/
func (i *dbInstance) GenerateService() (retVal api_v1.Service) {
	var fileContent []byte
	parsedData := api_v1.Service{}

	fileContent, err := ioutil.ReadFile(path.Join("/opt/servicebroker/templates", "service.json"))
	if err != nil {
		print(err)
	}
	glog.V(4).Infof("Debug: service.json: %s\n", fileContent)
	err = json.Unmarshal(fileContent, &parsedData)

	if err != nil {
		print(err)
	}

	parsedData.SetName("mysql-" + i.Params["cluster"].(string))
	var labels map[string]string
	labels = make(map[string]string)
	labels["app"] = "mysql-" + i.Params["cluster"].(string)
	parsedData.SetLabels(labels)
	parsedData.Spec.Selector = labels

	return parsedData
}

func getKubernetesClient(kubeConfigPath string) (clientset.Interface, error) {
	var clientConfig *clientrest.Config
	var err error
	if kubeConfigPath == "" {
		clientConfig, err = clientrest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeConfigPath)
		if err != nil {
			return nil, err
		}

		clientConfig, err = clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	return clientset.NewForConfig(clientConfig)
}

// Verify makes sure the db instance exists in the k8s cluster
func (b *BusinessLogic) Verify(i *dbInstance) bool {
	k8sClient, err := getKubernetesClient("")
	if err != nil {
		glog.V(4).Infof("can't create a client - PANIC")
		panic(err.Error())
	}
	getOpts := &meta_v1.GetOptions{}
	_, err = k8sClient.AppsV1().StatefulSets(b.dbNamespace).Get("mysql-"+i.Params["cluster"].(string), *getOpts)
	if err != nil {
		return false
	}

	return true
}
