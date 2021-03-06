package broker

import (
	"flag"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type Options struct {
	CatalogPath        string
	Async              bool
	DBConnectionString string
	EtcdEndpoints      string
	DBNamespace        string
	PiiDBNamespace     string
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func AddFlags(o *Options) {
	flag.StringVar(&o.CatalogPath, "catalogPath", "", "The path to the catalog")
	flag.BoolVar(&o.Async, "async", false, "Indicates whether the broker is handling the requests asynchronously.")
	flag.StringVar(&o.DBConnectionString, "db-connection", "root:password@tcp(mysql:3306)/broker", "database connection")
	flag.StringVar(&o.EtcdEndpoints, "etcd-endpoint", "etcd-cluster-client:2379", "etcd endpoint")
	flag.StringVar(&o.DBNamespace, "db-namespace", "test-ns", "namespace for db clusters")
	flag.StringVar(&o.PiiDBNamespace, "pii-db-namespace", "test-ns", "namespace for pii db clusters")
}
