package framework

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Framework is used in the execution of e2e tests.
type Framework struct {
	Client kubernetes.Interface
}

// New constructs a new e2e test Framework.
func New() *Framework { return &Framework{} }

// Init initialises the e2e test framework.
func (f *Framework) Init(kubeconfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	f.Client, err = kubernetes.NewForConfig(config)
	return err
}

// Run the tests and exit with the status code
func (f *Framework) Run(run func() int) {
	os.Exit(run())
}
