# Provision Flow
The A-Sync Flow of request to provision cluster

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
+-----------------------+        Pro^ision Req    +----------+------------+          +----------v-----------+          +--------------------+
|                       | +----------------------^+                       |          |                      |          |                    |
|                       |                         |     MySQL-Broker      |          |     Fullfilment      |          |      K8S           |
|   Instance Provision  |      Async ID           |                       |          |     Task Watcher     +---------^+      Stateful-Set  |
|                       +^------------------------+                       |          |                      |          |                    |
+----------+------------+                         +-----------------------+          +----------+-----------+          +--------------------+
           ^                                                                                    |
           |                                                                                    |
           |                                                                                    |
           |                                                                                    |
           |                            Confirmation                                            |
           +------------------------------------------------------------------------------------+


```

Implamantation tasks:
- [ ] Provision parameters
- [ ] Save order to ETCD
- [ ] Pull order from ETCD
- [ ] Create Stateful Set from an order
- [ ] Define creation permissions for the service
- [ ] Helm deployment
- [ ] Permissions for ETCD access
- [ ] A-Sync flow