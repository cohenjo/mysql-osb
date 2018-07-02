# Binding Flow
Binding creates the user/pasword and access details to access the cluster

```
                                                  +----------------------+
                                                  |                      |
                                                  |        ETCD          +----------------------+
                                                  |                      |                      |
                                                  +----------+-----------+                      |
                                                             ^                                  |
                                                             |                                  |
                                                             | Notify                           |
                                                             |                                  |
                                                             |                                  |
+-----------------------+       Bind              +----------+------------+          +----------v-----------+          +--------------------+
|                       | +----------------------^+                       |          |                      |          |                    |
|                       |                         |     MySQL+Broker      |          |     Fullfilment      |          |   K8S              |
|   Binding Request     |      recei^e secret     |                       |          |     Task Watcher     +---------^+   Secret           |
|                       +^------------------------+                       |          |                      |          |                    |
+-----------------------+                         +-----------------------+          +----------------------+          +--------------------+

```


Implamantation tasks:
- [ ] Binding parameters
- [ ] Save order to ETCD
- [ ] Pull order from ETCD
- [ ] Create Secret from Binding
- [ ] Define creation permissions for the service
- [ ] Helm deployment
- [ ] Permissions for ETCD access
- [ ] A-Sync flow