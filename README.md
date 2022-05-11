# andproxy

It's universal, not only http reverse proxy.

The concept of this app is hang one handler to one port and delegate all work to it. At the moment the handler can deny or accept access, forward requests to next step servers, forward requests depending request source ip, load balancing.

If this proxy not provide all the needs - nothing stopping to forward requests to another load-baalncer and build multi-level architecture. For example if you want to split requests by country.

Configuratin setting in yaml file.

```yml
#/etc/andproxy/config.yml
global:
  logFDir: /var/log/

listenPorts:
  tcp4 80:
    accept: 
    deny: $pool1
    servers: $pool1
    filters: $filter1
    toport: 80
    balancingMethod: roundRobin
  tcp4 2222:
    servers: 
      web1:
    toport: 22
    balancingMethod: none
  tcp4 3333:

filters:
  filter1:
    96.16.93.1: $pool1
  
    
serversPools:
  pool1: 
    web1:
    172.16.0.20:
      weight: 4
      maxFails: 3
      breakTime: 60
    172.16.0.30:
      weight: 2
      maxFails: 4
      breakTime: 120
pools:
  pool1: 
    - 111.54.12.0/24

  
  
```