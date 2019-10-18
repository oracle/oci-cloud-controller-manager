package nodepools

// APIOptionsV3 is a type to hold all of the node pools api options
type APIOptionsV3 struct {
	KubernetesVersions []string `json:"kubernetesVersions"`
	Images             []string `json:"images"`
	Shapes             []string `json:"shapes"`
}
