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
	"strings"
	"time"
)

var (
	numMapEntries = 100
	hzClient      *hazelcast.Client
	hzContext     = context.TODO()
)

type HzMetric struct {
	LastUpdated   time.Time
	Operation     string
	SourceIP      string
	InstanceIndex string
	RespTime      int64
}

func StartProbing() {
	hzConfig := getHZConfig()
	var err error
	if hzClient, err = hazelcast.StartNewClientWithConfig(hzContext, hzConfig); err != nil {
		fmt.Printf("failed to start Hazelcast hzClient: %s\n", err)
		os.Exit(8)
	} else {
		for {
			if !conf.StopRequested {
				startTime := time.Now()
				if hzMap, err := hzClient.GetMap(hzContext, conf.HzMapName); err != nil {
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
							getTime := time.Since(startTime).Microseconds()
							util.LogDebug(fmt.Sprintf("key \"%s\" found from IP %s with value: \"%s\" in %d microsecs\n", mapkey, conf.MyIP, value, getTime))
							// we did a successful get for a key, we now try a put
							value = fmt.Sprintf("value updated by %s at %s", conf.MyIP, time.Now().Format(time.RFC3339))
							startTime = time.Now()
							if err = hzMap.Set(hzContext, mapkey, value); err != nil {
								fmt.Printf("failed to Set value for key \"%s\": %s\n", mapkey, err)
							} else {
								putTime := time.Since(startTime).Microseconds()
								util.LogDebug(fmt.Sprintf("key \"%s\" updated with value: \"%s\" in %d microsecs\n", mapkey, value, putTime))
								SaveInHZ(hzContext, hzMap, getTime, putTime)
							}
						}
					}
				}
				time.Sleep(time.Duration(conf.IntervalSecs+rand.IntN(5)) * time.Second)
			}
		}
	}
}

// SaveInHZ - Store the getTime and putTime in hazelcast (as HzMetric structures)
func SaveInHZ(ctx context.Context, hzMap *hazelcast.Map, getTime int64, putTime int64) {
	mapValueGet := HzMetric{LastUpdated: time.Now(), Operation: "get", SourceIP: conf.MyIP, InstanceIndex: fmt.Sprintf("%d", conf.CFInstanceIndex), RespTime: getTime}
	if err := hzMap.Set(ctx, fmt.Sprintf("%s/%d/get", conf.CFEnv, conf.CFInstanceIndex), mapValueGet); err != nil {
		fmt.Printf("failed to store get metric: %s\n", err)
	}
	mapValuePut := HzMetric{LastUpdated: time.Now(), Operation: "put", SourceIP: conf.MyIP, InstanceIndex: fmt.Sprintf("%d", conf.CFInstanceIndex), RespTime: putTime}
	if err := hzMap.Set(ctx, fmt.Sprintf("%s/%d/put", conf.CFEnv, conf.CFInstanceIndex), mapValuePut); err != nil {
		fmt.Printf("failed to store put metric: %s\n", err)
	}
}

//GetMetricsFromHZ - Get the metrics from Hazelcast and store them in the Prometheus collector

func GetMetricsFromHZ() ([]HzMetric, error) {
	if hzMap, err := hzClient.GetMap(hzContext, conf.HzMapName); err != nil {
		return nil, fmt.Errorf("failed to get map %s: %w", conf.HzMapName, err)
	} else {
		if mapKeys, err := hzMap.GetKeySet(hzContext); err != nil {
			return nil, fmt.Errorf("failed to get keyset from map %s: %w", conf.HzMapName, err)
		} else {
			metrics := make([]HzMetric, 0, len(mapKeys))
			for _, mapKey := range mapKeys {
				if strings.HasPrefix(fmt.Sprintf("%s", mapKey), fmt.Sprintf("%s/", conf.CFEnv)) {
					if value, err := hzMap.Get(hzContext, mapKey); err != nil {
						fmt.Printf("failed to get value for key \"%s\": %s\n", mapKey, err)
					} else if metric, ok := value.(HzMetric); ok {
						metrics = append(metrics, metric)
					} else {
						fmt.Printf("value for key \"%s\" is not of type HzMetric\n", mapKey)
					}
				}
			}
			return metrics, nil
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
