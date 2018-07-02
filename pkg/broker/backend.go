package broker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	_ "github.com/go-sql-driver/mysql" // import the mysql driver
	"github.com/golang/glog"

	"k8s.io/api/apps/v1beta1"
	api_v1 "k8s.io/api/core/v1"
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

func (b *BusinessLogic) etcIt(request *osb.ProvisionRequest, i *dbInstance) {
	glog.V(4).Infof("storing InstanceL in etcd: %s !\n", request.InstanceID)
	key := fmt.Sprintf("mysql-broker/instance/%s", i.Params["cluster"].(string))
	value := "cluster info"

	glog.V(4).Infof("put in etcd:  %s, %s !\n", key, value)

}

func (b *BusinessLogic) order(request *osb.ProvisionRequest, i *dbInstance) {

	glog.V(4).Infof("InstanceL %s !\n", request.InstanceID)
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

	svc, err := k8sClient.CoreV1().Services("test-ns").Create(&ret)
	if err != nil {
		glog.V(4).Infof("can't create a service - PANIC")
		// panic(err.Error())
		glog.V(4).Infof("can't create a service - Don't panic it exists")
	}
	glog.V(4).Infof("Debug: service status %s\n", svc.Status.String())

	cfm := i.GenerateMySQLConfigMap()
	_, err = k8sClient.CoreV1().ConfigMaps("test-ns").Create(&cfm)
	if err != nil {
		glog.V(4).Infof("can't create a config map - PANIC, %s\n", err.Error())
		// fmt.Println("fuck")
		// panic(err.Error())
		glog.V(4).Infof("can't create a service - Don't panic it exists")
	}

	retss := i.GenerateStatefulSets()

	_, err = k8sClient.AppsV1beta1().StatefulSets("test-ns").Create(&retss)
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
	glog.V(4).Infof("######################################################################################################")
	glog.V(4).Infof(parsedData.String())
	glog.V(4).Infof("######################################################################################################")
	glog.V(4).Infof("######################################################################################################")
	return parsedData
}

/**
GenerateMySQLConfigMap generates the configuration map for a new cluster
*/
func (i *dbInstance) GenerateMySQLConfigMap() (retVal api_v1.ConfigMap) {
	var fileContent []byte
	parsedData := api_v1.ConfigMap{}

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
