package hz

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/rabobank/hzmon/conf"
	"github.com/rabobank/hzmon/util"
	"math/rand/v2"
	"os"
	"time"
)

var numMapEntries = 100

func StartProbing() {
	hzConfig := getHZConfig()
	hzContext := context.TODO()
	if client, err := hazelcast.StartNewClientWithConfig(hzContext, hzConfig); err != nil {
		fmt.Printf("failed to start Hazelcast client: %s\n", err)
		os.Exit(8)
	} else {
		for {
			if !conf.StopRequested {
				startTime := time.Now()
				if hzMap, err := client.GetMap(hzContext, conf.HzMapName); err != nil {
					fmt.Printf("failed to get map %s: %s\n", conf.HzMapName, err)
				} else {
					mapkey := fmt.Sprintf("testkey-%d", rand.IntN(100))
					if value, err := hzMap.Get(hzContext, mapkey); err != nil {
						fmt.Printf("failed to get value for key \"%s\": %s\n", err, mapkey)
					} else {
						if value == nil {
							//map is empty (probably TTL expired) or key not found, populate it
							fmt.Printf("key \"%s\" not found in map '%s', adding %d new keys now.\n", mapkey, conf.HzMapName, numMapEntries)
							for ix := 0; ix < numMapEntries; ix++ {
								value = fmt.Sprintf("value %d added by %s at %s", ix, conf.MyIP, time.Now().Format(time.RFC3339))
								if err = hzMap.Set(hzContext, fmt.Sprintf("testkey-%d", ix), value); err != nil {
									fmt.Printf("failed to set value for key \"testkey-%d\": %s\n", ix, err)
								}
							}
							fmt.Printf("added %d keys in %d microsecs\n", numMapEntries, time.Since(startTime).Microseconds())
						} else {
							conf.HzSliceLock.Lock()
							conf.HzGetTimes = append(conf.HzGetTimes, time.Since(startTime).Microseconds())
							util.LogDebug(fmt.Sprintf("key \"%s\" found from IP %s with value: \"%s\" in %d microsecs\n", mapkey, conf.MyIP, value, conf.HzGetTimes[len(conf.HzGetTimes)-1]))
							// we did a successful get for a key, we now try a put
							value = fmt.Sprintf("value updated by %s at %s", conf.MyIP, time.Now().Format(time.RFC3339))
							startTime = time.Now()
							if err = hzMap.Set(hzContext, mapkey, value); err != nil {
								fmt.Printf("failed to Set value for key \"%s\": %s\n", mapkey, err)
							} else {
								conf.HzPutTimes = append(conf.HzPutTimes, time.Since(startTime).Microseconds())
								util.LogDebug(fmt.Sprintf("key \"%s\" updated with value: \"%s\" in %d microsecs\n", mapkey, value, conf.HzPutTimes[len(conf.HzPutTimes)-1]))
							}
							conf.HzSliceLock.Unlock()
						}
					}
				}
				time.Sleep(time.Duration(conf.IntervalSecs+rand.IntN(5)) * time.Second)
			}
		}
	}
}

func getHZConfig() hazelcast.Config {

	connStrategy := cluster.ConnectionStrategyConfig{
		Retry:         cluster.ConnectionRetryConfig{InitialBackoff: types.Duration(100 * time.Millisecond), MaxBackoff: types.Duration(3 * time.Second), Multiplier: 2},
		Timeout:       types.Duration(2 * time.Second),
		ReconnectMode: cluster.ReconnectModeOff}

	clusterConfig := cluster.Config{
		Security:           cluster.SecurityConfig{Credentials: cluster.CredentialsConfig{Username: conf.HzConfigFromVCAP.Credentials.Principal, Password: conf.HzConfigFromVCAP.Credentials.Password}},
		Name:               conf.HzConfigFromVCAP.Credentials.ClusterName,
		Network:            cluster.NetworkConfig{Addresses: conf.HzConfigFromVCAP.Credentials.Ips},
		ConnectionStrategy: connStrategy,
	}

	failoverClusterConfig := cluster.Config{
		Security:           cluster.SecurityConfig{Credentials: cluster.CredentialsConfig{Username: conf.HzConfigFromVCAP.Credentials.Failover.Principal, Password: conf.HzConfigFromVCAP.Credentials.Failover.Password}},
		Name:               conf.HzConfigFromVCAP.Credentials.Failover.ClusterName,
		Network:            cluster.NetworkConfig{Addresses: conf.HzConfigFromVCAP.Credentials.Failover.Ips},
		ConnectionStrategy: connStrategy,
	}

	failoverClusterConfigs := []cluster.Config{clusterConfig, failoverClusterConfig}

	return hazelcast.Config{ClientName: "panzer-hzmon", Failover: cluster.FailoverConfig{Configs: failoverClusterConfigs, TryCount: 1, Enabled: true}}
}
