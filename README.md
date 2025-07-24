### HzMon - Hazelcast Monitor

An app that reads from and writes to a configured Hazelcast cluster at specified intervals. 
It exports prometheus metrics (response time) at /metrics. 
It also offers a very simple web interface to stop/start the hz probing and turn debug on/off, to reach this interface for each app instance (i.e. turn debug on):
If you want to do that for all app instances: 
```shell
# target org and space where hzmon is running
cf t -o <org> -s <space>
# get the process GUID of hzmon
PGUID=$(cf curl /v3/apps/"$(cf app hzmon --guid)"/processes | jq -r '.resources[].guid')
# iterate over all instances of hzmon and turn debug on
for IX in $(seq 0 <number of instances - 1>);
do
curl -s -H "X-Cf-Process-Instance: $PGUID:$IX" "https://hzmon.<cf apps domain>?debugoff"
done
```

The response time metrics are also stored in hazelcast, so prometheus can call any instance of hzmon to get the metrics.  
Sample prometheus metrics:

```
rabo_hzmon_responsetime{instanceIndex="3",operation="get",sourceIP="10.253.21.52"} 3379
rabo_hzmon_responsetime{instanceIndex="3",operation="put",sourceIP="10.253.21.52"} 1746
rabo_hzmon_responsetime{instanceIndex="4",operation="get",sourceIP="10.253.21.55"} 1787
rabo_hzmon_responsetime{instanceIndex="4",operation="put",sourceIP="10.253.21.55"} 853
```
