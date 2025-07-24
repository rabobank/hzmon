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
	VERSION string
)

const (
	IntervalSecsDefault = 60
	HzMapNameDefault    = "it4it-org.panzer-space.test-cache.test-cache"
)

var (
	intervalSecsStr    = os.Getenv("INTERVAL_SECS")
	IntervalSecs       int
	debugStr           = os.Getenv("DEBUG")
	Debug              bool
	MyIP               string
	StopRequested      bool
	HzConfigFromVCAP   model.UserProvided
	HzMapName          = os.Getenv("HZ_MAP_NAME")
	CFEnv              = os.Getenv("RABOPCF_SYSTEM_ENV")
	cfInstanceIndexStr = os.Getenv("CF_INSTANCE_INDEX")
	CFInstanceIndex    int
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
