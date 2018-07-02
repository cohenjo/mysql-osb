
# Deploying Orch for this crazy thingy




kubectl create configmap orch-config --from-file=/Users/jony/Documents/mysql/orchestrator/orchestrator.conf.json -n mysql-broker
kubectl create -f /Users/jony/Documents/mysql/orchestrator/orcg.yaml


on all mysql clusters run:

CREATE USER 'orchestrator'@'%' IDENTIFIED BY '0rch3strat0r1';
GRANT SUPER, PROCESS, REPLICATION SLAVE, REPLICATION CLIENT, RELOAD ON *.* TO 'orchestrator'@'%';
GRANT SELECT ON meta.* TO 'orchestrator'@'%';
