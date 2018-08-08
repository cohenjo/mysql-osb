package broker

import (
	"database/sql"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type Parameters struct {
	Artifact       string
	DeploymentType string
	Size           int
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

func (b *BusinessLogic) mysqlStore(request *osb.ProvisionRequest, i *dbInstance) {

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
}
