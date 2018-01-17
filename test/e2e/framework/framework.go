package framework

import (
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Poll defines how regularly to poll kubernetes resources.
	Poll = 2 * time.Second
)

// Framework is used in the execution of e2e tests.
type Framework struct {
	Client kubernetes.Interface

	// Namespace in which test resources are created.
	Namespace string
}

// New constructs a new e2e test Framework.
func New() *Framework { return &Framework{} }

// Init initialises the e2e test framework.
func (f *Framework) Init(kubeconfig, namespace string) error {
	f.Namespace = namespace

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
