
# Deploying Orch for this crazy thingy




kubectl create configmap orch-config --from-file=/Users/jony/Documents/mysql/orchestrator/orchestrator.conf.json -n mysql-broker
kubectl create -f /Users/jony/Documents/mysql/orchestrator/orcg.yaml

