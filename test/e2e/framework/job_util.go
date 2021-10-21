package framework

import (
	"context"
	"fmt"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

var ErrJobFailed = fmt.Errorf("Job failed")

//Creates a new job which will run a pod with the centos container running the given script
func (j *ServiceTestJig) CreateJobRunningScript(ns string, script string, backOffLimit int32, name string){
	job, err := j.Client.BatchV1().Jobs(ns).Create(context.Background(), &v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind: "Job",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: ns,
		},
		Spec: v1.JobSpec{
			Template: v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name: name,
							Image: centos,
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", script},
						},
					},
					RestartPolicy: v12.RestartPolicyOnFailure,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	},metav1.CreateOptions{})
	if err!= nil{
		Failf("Error creating job: %v", err)
	}
	err = j.waitTimeoutForJobCompletedInNamespace(job.Name, ns, JobCompletionTimeout)
	if err != nil {
		Failf("Job %q did not complete: %v", job.Name, err)
	}
}

func (j *ServiceTestJig) waitTimeoutForJobCompletedInNamespace(jobName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.jobCompleted(jobName, namespace))
}

func (j *ServiceTestJig) jobCompleted(jobName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		job, err := j.Client.BatchV1().Jobs(namespace).Get(context.Background(),jobName,metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if job.Status.Succeeded == 1{
			return true, nil
		}
		if job.Status.Failed == 1 {
			return false, ErrJobFailed
		}
		return false, nil
	}
}
