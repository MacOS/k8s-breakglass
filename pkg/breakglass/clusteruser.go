package breakglass

type ClusterUserGroup struct {
	Clustername string `json:"clustername,omitempty"`
	Username    string `json:"username,omitempty"`
	Groupname   string `json:"clustergroup,omitempty"`
}
