package broker

import (
	"github.com/cohenjo/mysql-osb/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// redis-operator/api/redisfailover/v1alpha2/types.go
// func getResources(spec redisfailoverv1alpha2.RedisFailoverSpec) corev1.ResourceRequirements {
// 	return corev1.ResourceRequirements{
// 		Requests: getRequests(spec.Redis.Resources),
// 		Limits:   getLimits(spec.Redis.Resources),
// 	}
// }

func getLimits(resources types.MySQLResources) corev1.ResourceList {
	return generateResourceList(resources.Limits.CPU, resources.Limits.Memory)
}

func getRequests(resources types.MySQLResources) corev1.ResourceList {
	return generateResourceList(resources.Requests.CPU, resources.Requests.Memory)
}

func generateResourceList(cpu string, memory string) corev1.ResourceList {
	resources := corev1.ResourceList{}
	if cpu != "" {
		resources[corev1.ResourceCPU], _ = resource.ParseQuantity(cpu)
	}
	if memory != "" {
		resources[corev1.ResourceMemory], _ = resource.ParseQuantity(memory)
	}
	return resources
}

func generateStorageRequest(storage string) corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{}
	if storage != "" {
		resources.Requests[corev1.ResourceRequestsStorage], _ = resource.ParseQuantity(storage)
	}
	return resources
}
