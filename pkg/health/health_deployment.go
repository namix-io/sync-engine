package health

import (
	"fmt"

	"github.com/namix-io/sync-engine/pkg/utils/kube"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func getDeploymentHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	gvk := obj.GroupVersionKind()
	switch gvk {
	case appsv1.SchemeGroupVersion.WithKind(kube.DeploymentKind):
		var deployment appsv1.Deployment
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured Deployment to typed: %v", err)
		}
		return getAppsv1DeploymentHealth(&deployment)
	case appsv1beta1.SchemeGroupVersion.WithKind(kube.DeploymentKind):
		var deployment appsv1beta1.Deployment
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured Deployment to typed: %v", err)
		}
		return getAppsv1beta1DeploymentHealth(&deployment)
	case extv1beta1.SchemeGroupVersion.WithKind(kube.DeploymentKind):
		var deployment extv1beta1.Deployment
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured Deployment to typed: %v", err)
		}
		return getExtv1beta1DeploymentHealth(&deployment)
	default:
		return nil, fmt.Errorf("unsupported Deployment GVK: %s", gvk)
	}
}

func getAppsv1DeploymentHealth(deployment *appsv1.Deployment) (*HealthStatus, error) {
	if deployment.Spec.Paused {
		return &HealthStatus{
			Status:  HealthStatusSuspended,
			Message: "Deployment is paused",
		}, nil
	}
	// Borrowed at kubernetes/kubectl/rollout_status.go https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L80
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := getAppsv1DeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
		if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
			return &HealthStatus{
				Status:  HealthStatusDegraded,
				Message: fmt.Sprintf("Deployment %q exceeded its progress deadline", deployment.Name),
			}, nil
		} else if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...", deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas),
			}, nil
		} else if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d old replicas are pending termination...", deployment.Status.Replicas-deployment.Status.UpdatedReplicas),
			}, nil
		} else if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas),
			}, nil
		}
	} else {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Waiting for rollout to finish: observed deployment generation less then desired generation",
		}, nil
	}

	return &HealthStatus{
		Status: HealthStatusHealthy,
	}, nil
}

func getAppsv1beta1DeploymentHealth(deployment *appsv1beta1.Deployment) (*HealthStatus, error) {
	if deployment.Spec.Paused {
		return &HealthStatus{
			Status:  HealthStatusSuspended,
			Message: "Deployment is paused",
		}, nil
	}
	// Borrowed at kubernetes/kubectl/rollout_status.go https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L80
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := getAppsv1beta1DeploymentCondition(deployment.Status, appsv1beta1.DeploymentProgressing)
		if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
			return &HealthStatus{
				Status:  HealthStatusDegraded,
				Message: fmt.Sprintf("Deployment %q exceeded its progress deadline", deployment.Name),
			}, nil
		} else if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...", deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas),
			}, nil
		} else if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d old replicas are pending termination...", deployment.Status.Replicas-deployment.Status.UpdatedReplicas),
			}, nil
		} else if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas),
			}, nil
		}
	} else {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Waiting for rollout to finish: observed deployment generation less then desired generation",
		}, nil
	}

	return &HealthStatus{
		Status: HealthStatusHealthy,
	}, nil
}

func getExtv1beta1DeploymentHealth(deployment *extv1beta1.Deployment) (*HealthStatus, error) {
	if deployment.Spec.Paused {
		return &HealthStatus{
			Status:  HealthStatusSuspended,
			Message: "Deployment is paused",
		}, nil
	}
	// Borrowed at kubernetes/kubectl/rollout_status.go https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L80
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := getExtv1beta1DeploymentCondition(deployment.Status, extv1beta1.DeploymentProgressing)
		if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
			return &HealthStatus{
				Status:  HealthStatusDegraded,
				Message: fmt.Sprintf("Deployment %q exceeded its progress deadline", deployment.Name),
			}, nil
		} else if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...", deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas),
			}, nil
		} else if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d old replicas are pending termination...", deployment.Status.Replicas-deployment.Status.UpdatedReplicas),
			}, nil
		} else if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas),
			}, nil
		}
	} else {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Waiting for rollout to finish: observed deployment generation less then desired generation",
		}, nil
	}

	return &HealthStatus{
		Status: HealthStatusHealthy,
	}, nil
}

func getAppsv1DeploymentCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
func getAppsv1beta1DeploymentCondition(status appsv1beta1.DeploymentStatus, condType appsv1beta1.DeploymentConditionType) *appsv1beta1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func getExtv1beta1DeploymentCondition(status extv1beta1.DeploymentStatus, condType extv1beta1.DeploymentConditionType) *extv1beta1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
