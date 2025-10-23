package kubernetes

// Route represents the OpenShift route object structure
type Route struct {
	Spec struct {
		Host string `json:"host"`
	} `json:"spec"`
}
