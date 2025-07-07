package conf

import (
	"encoding/json"
	"fmt"
	"github.com/rabobank/hzmon/model"
	"log"
	"os"
	"strconv"
)

var (
	CommitHash string
	BuildTime  string
)

const (
	IntervalSecsDefault            = 60
	PushGatewayIntervalSecsDefault = 60
	HzMapNameDefault               = "it4it-org.panzer-space.test-cache.test-cache"
)

var (
	intervalSecsStr            = os.Getenv("INTERVAL_SECS")
	IntervalSecs               int
	pushGatewayIntervalSecsStr = os.Getenv("PUSH_INTERVAL_SECS")
	PushGatewayIntervalSecs    int
	debugStr                   = os.Getenv("DEBUG")
	Debug                      bool
	MyIP                       string
	StopRequested              bool
	HzConfigFromVCAP           model.UserProvided
	HzMapName                  = os.Getenv("HZ_MAP_NAME")
	cfInstanceIndexStr         = os.Getenv("CF_INSTANCE_INDEX")
	CFInstanceIndex            int
	HzGetTime                  int64 // time in microseconds it took to get the hz-map entry
	PushGatewayURL             = os.Getenv("PUSHGATEWAY_URL")
)

func EnvironmentComplete() bool {
	envComplete := true
	if intervalSecsStr == "" {
		IntervalSecs = IntervalSecsDefault
	} else {
		var err error
		IntervalSecs, err = strconv.Atoi(intervalSecsStr)
		if err != nil {
			log.Printf("failed to parse INTERVAL_SECS: %s", err)
			envComplete = false
		}
	}

	if pushGatewayIntervalSecsStr == "" {
		PushGatewayIntervalSecs = PushGatewayIntervalSecsDefault
	} else {
		var err error
		PushGatewayIntervalSecs, err = strconv.Atoi(pushGatewayIntervalSecsStr)
		if err != nil {
			log.Printf("failed to parse PUSH_INTERVAL_SECS: %s", err)
			envComplete = false
		}
	}

	if cfInstanceIndexStr == "" {
		CFInstanceIndex = 0
	} else {
		var err error
		CFInstanceIndex, err = strconv.Atoi(cfInstanceIndexStr)
		if err != nil {
			log.Printf("failed to parse CF_INSTANCE_INDEX: %s", err)
			envComplete = false
		}
	}

	if PushGatewayURL == "" {
		fmt.Println("missing envvar : PUSHGATEWAY_URL, not sending metrics to pushgateway")
	}

	if HzMapName == "" {
		HzMapName = HzMapNameDefault
	}

	if !IsHZComplete() {
		envComplete = false
	}

	if debugStr == "true" {
		Debug = true
	}

	return envComplete
}

func IsHZComplete() bool {
	vcapServicesString := os.Getenv("VCAP_SERVICES")
	if vcapServicesString != "" {
		vcapServices := model.VcapServices{}
		if err := json.Unmarshal([]byte(vcapServicesString), &vcapServices); err != nil {
			fmt.Printf("could not get hazelcast credentials from user-provided service, error: %s\n", err)
		} else {
			for _, service := range vcapServices.UserProvided {
				if service.InstanceName == "test-cache" {
					fmt.Printf("got hz-credentials, instance-name:%s, clustername:%s, failover-clustername:%s,  \n", service.InstanceName, service.Credentials.ClusterName, service.Credentials.Failover.ClusterName)
					HzConfigFromVCAP = service
					return true
				}
			}
		}
	}
	fmt.Println("no VCAP_SERVICES envvar found")
	return false
}
