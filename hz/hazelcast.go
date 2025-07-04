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

func StartProbing() {
	hzConfig := getHZConfig()
	hzContext := context.TODO()
	if client, err := hazelcast.StartNewClientWithConfig(hzContext, hzConfig); err != nil {
		fmt.Printf("failed to start Hazelcast client: %s\n", err)
		os.Exit(8)
	} else {
		for {
			startTime := time.Now()
			if getMap, err := client.GetMap(hzContext, conf.HzMapName); err != nil {
				fmt.Printf("failed to get map %s: %s\n", conf.HzMapName, err)
			} else {
				if value, err := getMap.Get(hzContext, "testKey"); err != nil {
					fmt.Printf("failed to get value for key \"testKey\": %s\n", err)
				} else {
					getTime := time.Since(startTime)
					if value == nil {
						fmt.Printf("Key \"testKey\" not found in map '%s', adding it now.\n", conf.HzMapName)
						startTime = time.Now()
						value = fmt.Sprintf("value added by %s at %s", conf.MyIP, time.Now().Format(time.RFC3339))
						if err = getMap.Set(hzContext, "testKey", value); err != nil {
							fmt.Printf("failed to set value for key \"testKey\": %s\n", err)
						} else {
							setTime := time.Since(startTime)
							util.LogDebug(fmt.Sprintf("Key \"testKey\" added to map '%s' with value \"%s\" in %d microsecs\n", conf.HzMapName, value, setTime.Microseconds()))
						}
					} else {
						util.LogDebug(fmt.Sprintf("Key \"testKey\" found from IP %s with value: \"%s\" in %d microsecs\n", conf.MyIP, value, getTime.Microseconds()))
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
