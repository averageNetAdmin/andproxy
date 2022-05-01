# andproxy

It's universal, not only http reverse-proxy. You can listen any port and handle all requests. 

Configuratin setting in yaml file.

```yml
#/etc/andproxy/config.yml
listenPorts:
  tcp4 80:
    accept: 
    deny: $pool1      #Use addressPools pool
    servers: $pool1   #Use servers pool
    filter: $filter1
    toport: 80
  tcp4 443:

  tcp4 3333:

staticFilters:
  filter1:
    192.168.1.1: $poolname      #Use servers pool
    $poolname: 172.16.0.20      #Use addressPools pool
    192.168.100.2/24: 172.16.0.30
servers:
  poolname: 
    - 172.16.0.10
    - 172.16.0.20
    - 172.16.0.30
addressPools:
  poolname: 
    - 192.16.0.10
    - 192.16.0.20
    - 192.16.0.30
  
```
