

# Setup
minikube start  --memory=2048 --cpus=4 --hyperv-virtual-switch="primary-virtual-switch" --v=7 --alsologtostderr --bootstrapper=kubeadm --extra-config=apiserver.authorization-mode=RBAC

<!-- kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
minikube dashboard -->

## setup rbac
<!-- kubectl create serviceaccount tiller --namespace kube-system -->
kubectl create -f DB_kube/rbac-config.yaml
helm init --upgrade 
<!-- helm init --upgrade --tiller-tls-verify -->

kubectl create clusterrolebinding tiller-cluster-admin \
    --clusterrole=cluster-admin \
    --serviceaccount=kube-system:default

configure 2 context to spin up pods in 96 & 42:
kubectl config set-context minikubeBroker --namespace=broker-skeleton \
  --cluster=minikube \
  --user=minikube

kubectl config set-context minikubeTest --namespace=test-ns \
  --cluster=minikube \
  --user=minikube


kubectl config current-context
kubectl config use-context minikubeBroker


helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com

helm search service-catalog

helm install svc-cat/catalog \
    --name catalog --namespace catalog



<!-- curl -sLO https://download.svcat.sh/cli/latest/darwin/amd64/svcat
chmod +x ./svcat
mv ./svcat /usr/local/bin/
 -->
svcat version --client

This will need a DB - will be create with Helm! 
 <!-- kubectl create -f DB_kube/mysql-single.yaml -->
( kubectl run -it --rm --image=mysql:5.6 --restart=Never mysql-client -- mysql -h mysql.mysql-broker -ppassword)

in ~/Documents/mysql/orchestrator/deployment ==> deploy orchestrator

# Develop


## consider this instead: https://github.com/kubernetes-incubator/service-catalog/blob/master/contrib/pkg/broker/user_provided/controller/controller.go ??

 go get github.com/cohenjo/mysql-osb
 cd $GOPATH/src/github.com/cohenjo/mysql-osb


<!-- IMAGE=cohenjo/broker TAG=latest make push deploy-helm -->
IMAGE=cohenjo/broker TAG=latest PULL=Always make push deploy-helm

svcat get brokers
svcat describe broker mysql-broker

kubectl get clusterservicebrokers broker-skeleton -o yaml

svcat get classes
svcat describe class mysql-artifact-service

svcat get plans
svcat describe plan mysql-artifact-service/default


# test the service

kubectl create namespace test-ns
kubectl create -f DB_kube/db_broker/mysql_instance.yaml

svcat describe instance -n test-ns mysql-instance

kubectl create -f DB_kube/db_broker/mysql-binding.yaml
svcat describe binding -n test-ns mysql-binding

kubectl get secrets -n test-ns

https://broker-skeleton-broker-skeleton.broker-skeleton.svc.cluster.local

# Cleanup
svcat unbind -n test-ns mysql-instance
<!-- kubectl delete -n test-ns servicebindings mysql-binding -->

svcat deprovision -n test-ns mysql-instance
<!-- kubectl delete -n test-ns serviceinstances mysql-instance -->

kubectl delete -n mysql-broker  svc mysql
kubectl delete -n mysql-broker  pvc mysql-broker-pvclaim


kubectl delete clusterservicebrokers mysql-broker
helm delete --purge mysql-broker
kubectl delete ns mysql-broker
kubectl delete ns test-ns








to allow no permissions you can use:
-----------------------------------
kubectl create clusterrolebinding permissive-binding \
  --clusterrole=cluster-admin \
  --user=admin \
  --user=kubelet \
  --group=system:serviceaccounts