package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/golang/glog"
	"k8s.io/api/apps/v1beta1"
	api_v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	clientrest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cohenjo/mysql-osb/pkg/broker"
)

var options struct {
	broker.Options

	Port                 int
	Insecure             bool
	TLSCert              string
	TLSKey               string
	TLSCertFile          string
	TLSKeyFile           string
	AuthenticateK8SToken bool
	KubeConfig           string
}

func init() {
	flag.IntVar(&options.Port, "port", 8443, "use '--port' option to specify the port for broker to listen on")
	flag.BoolVar(&options.Insecure, "insecure", false, "use --insecure to use HTTP vs HTTPS.")
	flag.StringVar(&options.TLSCertFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	flag.StringVar(&options.TLSKeyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	flag.StringVar(&options.TLSCert, "tlsCert", "", "base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used.")
	flag.StringVar(&options.TLSKey, "tlsKey", "", "base-64 encoded PEM block to use as the private key matching the TLS certificate.")
	flag.BoolVar(&options.AuthenticateK8SToken, "authenticate-k8s-token", false, "option to specify if the broker should validate the bearer auth token with kubernetes")
	flag.StringVar(&options.KubeConfig, "kube-config", "", "specify the kube config path to be used")
	broker.AddFlags(&options.Options)
	flag.Parse()
}

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		glog.Fatalln(err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), "0.1.0")
		return nil
	}
	if (options.TLSCert != "" || options.TLSKey != "") &&
		(options.TLSCert == "" || options.TLSKey == "") {
		fmt.Println("To use TLS with specified cert or key data, both --tlsCert and --tlsKey must be used")
		return nil
	}

	// addr := ":" + strconv.Itoa(options.Port)

	k8sClient, err := getKubernetesClient(options.KubeConfig)
	if err != nil {
		return err
	}

	ret := GenerateHelloService()
	fmt.Println(ret.GetName())
	ret.SetName("others-mysql")
	fmt.Println(ret.GetName())
	k8sClient.CoreV1().Services("test-ns").Create(&ret)

	// if options.AuthenticateK8SToken {
	// 	// get k8s client
	// 	k8sClient, err := getKubernetesClient(options.KubeConfig)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// Create a User Info Authorizer.
	// 	authz := middleware.SARUserInfoAuthorizer{
	// 		SAR: k8sClient.AuthorizationV1().SubjectAccessReviews(),3
	// 	}
	// 	// create TokenReviewMiddleware
	// 	tr := middleware.TokenReviewMiddleware{
	// 		TokenReview: k8sClient.Authentication().TokenReviews(),
	// 		Authorizer:  authz,
	// 	}
	// 	// Use TokenReviewMiddleware.
	// 	s.Router.Use(tr.Middleware)
	// }

	glog.Infof("Starting broker!")

	return nil
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

func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			glog.Infof("Received SIGTERM, exiting gracefully...")
			f()
			os.Exit(0)
		case <-ctx.Done():
			os.Exit(0)
		}
	}
}

/*
GenerateHelloDeployment generates something
*/
func GenerateHelloDeployment() (retVal v1beta1.StatefulSet) {
	var fileContent []byte
	parsedData := v1beta1.StatefulSet{}

	fileContent, err := ioutil.ReadFile(path.Join("templates", "hello.json"))
	if err != nil {
		print(err)
	}
	err = json.Unmarshal(fileContent, &parsedData)

	if err != nil {
		print(err)
	}

	return parsedData
}

/*
GenerateHelloDeployment generates something
*/
func GenerateHelloService() (retVal api_v1.Service) {
	var fileContent []byte
	parsedData := api_v1.Service{}

	fileContent, err := ioutil.ReadFile(path.Join("templates", "service.json"))
	if err != nil {
		print(err)
	}
	err = json.Unmarshal(fileContent, &parsedData)

	if err != nil {
		print(err)
	}

	return parsedData
}
