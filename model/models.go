package model

type VcapServices struct {
	UserProvided []UserProvided `json:"user-provided"`
}

type UserProvided struct {
	BindingGUID string `json:"binding_guid"`
	BindingName any    `json:"binding_name"`
	Credentials struct {
		ClusterName string `json:"cluster-name"`
		Failover    struct {
			ClusterName string   `json:"cluster-name"`
			Ips         []string `json:"ips"`
			Password    string   `json:"password"`
			Principal   string   `json:"principal"`
		} `json:"failover"`
		Ips       []string `json:"ips"`
		Password  string   `json:"password"`
		Principal string   `json:"principal"`
	} `json:"credentials"`
	InstanceGUID   string   `json:"instance_guid"`
	InstanceName   string   `json:"instance_name"`
	Label          string   `json:"label"`
	Name           string   `json:"name"`
	SyslogDrainURL any      `json:"syslog_drain_url"`
	Tags           []string `json:"tags"`
	VolumeMounts   []any    `json:"volume_mounts"`
}
