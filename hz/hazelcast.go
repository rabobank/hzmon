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
		// initial wait time so that with many instances the activity will be spread out
		waitTime := float64(conf.IntervalSecs) / (1 + float64(conf.CFInstanceIndex))
		fmt.Printf("waiting %.0f seconds\n", waitTime)
		time.Sleep(time.Duration(waitTime) * time.Second)

		for {
			startTime := time.Now()
			if getMap, err := client.GetMap(hzContext, conf.HzMapName); err != nil {
				fmt.Printf("failed to get map %s: %s\n", conf.HzMapName, err)
			} else {
				mapKey := fmt.Sprintf("testKey-%d", rand.IntN(100))
				if value, err := getMap.Get(hzContext, mapKey); err != nil {
					fmt.Printf("failed to get value for key \"%s\": %s\n", err, mapKey)
				} else {
					if value == nil {
						//map is empty (probably TTL expired) or key not found, populate it
						fmt.Printf("Key \"%s\" not found in map '%s', adding %d new keys now.\n", mapKey, conf.HzMapName, numMapEntries)
						for ix := 0; ix < numMapEntries; ix++ {
							value = fmt.Sprintf("value %d added by %s at %s", ix, conf.MyIP, time.Now().Format(time.RFC3339))
							if err = getMap.Set(hzContext, fmt.Sprintf("testKey-%d", ix), value); err != nil {
								fmt.Printf("failed to set value for key \"testKey-%d\": %s\n", ix, err)
							}
						}
						fmt.Printf("added %d keys in %d microsecs\n", numMapEntries, time.Since(startTime).Microseconds())
					} else {
						util.LogDebug(fmt.Sprintf("Key \"%s\" found from IP %s with value: \"%s\" in %d microsecs\n", mapKey, conf.MyIP, value, time.Since(startTime).Microseconds()))
					}
				}
			}
			time.Sleep(time.Duration(conf.IntervalSecs+rand.IntN(5)) * time.Second)
		}
	}
}

func getHZConfig() hazelcast.Config {

	connStrategy := cluster.ConnectionStrategyConfig{Retry: cluster.ConnectionRetryConfig{
		InitialBackoff: types.Duration(100 * time.Millisecond),
		MaxBackoff:     types.Duration(10 * time.Second),
		Multiplier:     2,
		Jitter:         3},
		Timeout:       types.Duration(3 * time.Second),
		ReconnectMode: cluster.ReconnectModeOn}

	clusterConfig := cluster.Config{
		Security: cluster.SecurityConfig{
			Credentials: cluster.CredentialsConfig{Username: conf.HzConfigFromVCAP.Credentials.Principal, Password: conf.HzConfigFromVCAP.Credentials.Password}},
		Name: conf.HzConfigFromVCAP.Credentials.ClusterName,
		Network: cluster.NetworkConfig{Addresses: conf.HzConfigFromVCAP.Credentials.Ips,
			ConnectionTimeout: types.Duration(3 * time.Second),
		},
		ConnectionStrategy: connStrategy,
	}

	failoverClusterConfig1 := cluster.Config{
		Security: cluster.SecurityConfig{
			Credentials: cluster.CredentialsConfig{Username: conf.HzConfigFromVCAP.Credentials.Failover.Principal, Password: conf.HzConfigFromVCAP.Credentials.Failover.Password}},
		Name: conf.HzConfigFromVCAP.Credentials.Failover.ClusterName,
		Network: cluster.NetworkConfig{Addresses: conf.HzConfigFromVCAP.Credentials.Failover.Ips,
			ConnectionTimeout: types.Duration(3 * time.Second),
		},
		ConnectionStrategy: connStrategy,
	}

	failoverClusterConfigs := []cluster.Config{failoverClusterConfig1}

	return hazelcast.Config{ClientName: "panzer-hzmon",
		Failover: cluster.FailoverConfig{Configs: failoverClusterConfigs},
		Cluster:  clusterConfig,
		//Stats:    hazelcast.StatsConfig{Enabled: true, Period: types.Duration(10 * time.Second)}
	}
}
