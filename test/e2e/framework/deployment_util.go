package framework

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/pointer"
	"time"
)

func (j *PVCTestJig) createDeploymentOnNodeAndWait(command string, pvcName string, ns string, name string, replicas int32, nodeSelectorLabels map[string]string) string{

	deployment, err := j.KubeClient.AppsV1().Deployments(ns).Create(context.Background(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    name,
							Image:   centos,
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", command},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "persistent-storage",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "persistent-storage",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
					NodeSelector: nodeSelectorLabels,

				},
			},
		},
	}, metav1.CreateOptions{})

	if err != nil{
		Failf("Error creating deployment %v: %v", name, err)
	}

	// Waiting for deployment to be completed
	Logf("Waiting up to %v for deployment %v to be completed", deploymentAvailableTimeout, deployment.Name)
	err = j.waitTimeoutForDeploymentAvailable(deployment.Name, ns, deploymentAvailableTimeout, replicas)
	if err != nil {
		Failf("Deployment %q did not complete: %v", deployment.Name, err)
	}

	return deployment.Name
}

// waitTimeoutForDeploymentCompleted waits default amount of time (deploymentCompletionTimeout) for the specified deployment to complete
//Returns an error if timeout occurs first, or pod goes in to failed state.
func (j *PVCTestJig) waitTimeoutForDeploymentAvailable(deploymentName string, namespace string, timeout time.Duration, replicas int32) error {
	return wait.PollImmediate(Poll, timeout, j.deploymentAvailable(deploymentName, namespace, replicas))
}

func (j *PVCTestJig) deploymentAvailable(deploymentName string, namespace string, replicas int32) wait.ConditionFunc {
	return func() (bool, error) {
		deployment, err := j.KubeClient.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if deployment.Status.AvailableReplicas == replicas {
			return true, nil
		}
		return false, nil
	}
}
