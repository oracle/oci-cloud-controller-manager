package types

const (
	BMCLBShape100Mbps = "100Mbps"
	BMCLBShape400Mbps = "400Mbps"
)

// BMCLBBackendSet defines one backend set info for a BMC LB
type BMCLBBackendSet struct {
	BackendSetName string         `json:"backendSetName"`
	IngressPort    int            `json:"ingressPort"`
	Backends       []BMCLBBackend `json:"backends"`
}

// BMCLBBackend defines the backend info for a BMC LB
type BMCLBBackend struct {
	IPAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
	Weight    int    `json:"weight"`
}
