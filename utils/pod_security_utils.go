package utils

import corev1 "k8s.io/api/core/v1"

func securePodSpec(spec *corev1.PodSpec) {
	spec.AutomountServiceAccountToken = boolPtr(false)
	spec.SecurityContext = &corev1.PodSecurityContext{
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

func secureContainer(container *corev1.Container) {
	container.SecurityContext = &corev1.SecurityContext{
		AllowPrivilegeEscalation: boolPtr(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	}
}

func secureContainers(containers []corev1.Container) {
	for i := range containers {
		secureContainer(&containers[i])
	}
}

func SecurePodSpec(spec *corev1.PodSpec) {
	securePodSpec(spec)
	secureContainers(spec.InitContainers)
	secureContainers(spec.Containers)
}
