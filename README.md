### HzMon - Hazelcast Monitor

An app that reads from and writes to a configured Hazelcast cluster at specified intervals. 
Metrics (response time) are collected and sent to a prometheus push gateway.
It also offers a very simple web interface to stop/start the hz probing and turn debug on/off, to reach this interface for each app instance (i.e. turn debug on):

```
for IX in $(cf curl /v3/processes/$(cf app hzmon --guid)/stats|jq -r '.resources[].instance_guid')
do
curl -s -H "X-Cf-App-Instance=$IX" 'https://hzmon.<apps domain>/?debugon'
done
```
