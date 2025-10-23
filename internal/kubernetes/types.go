package kubernetes

// Route represents the OpenShift route object structure
type route struct {
	Spec struct {
		Host string `json:"host"`
	} `json:"spec"`
}
