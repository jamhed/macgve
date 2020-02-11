package mutate

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func mutateContainer(c corev1.Container, gve *GoVaultEnv) corev1.Container {
	args := append(c.Command, c.Args...)
	c.Command = []string{"/vault/govaultenv", "-addr=" + gve.vaultaddr, "-kubeauth=" + gve.authpath, "-stripname=true"}
	c.Args = args
	c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{Name: "govaultenv", MountPath: "/vault"})
	c.Env = append(c.Env, []corev1.EnvVar{
		{Name: "SSL_CERT_FILE", Value: "/vault/ca-certificates.crt"},
	}...)
	return c
}

func insertGve(pod *corev1.Pod, gve *GoVaultEnv) *corev1.Pod {
	for i, c := range pod.Spec.Containers {
		if gve.IsIn(c.Name) {
			pod.Spec.Containers[i] = mutateContainer(c, gve)
		}
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, makeInitContainer(gve.image))
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name:         "govaultenv",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}},
	})
	return pod
}

func makeInitContainer(gveImage string) corev1.Container {
	return corev1.Container{
		Name:            "govaultenv-init",
		Image:           gveImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"sh", "-c", "cp govaultenv /vault/ && cp ca-certificates.crt /vault/ca-certificates.crt"},
		VolumeMounts:    []corev1.VolumeMount{{Name: "govaultenv", MountPath: "/vault"}},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceLimitsCPU:    resource.MustParse("1"),
				corev1.ResourceLimitsMemory: resource.MustParse("500Mi")},
			Requests: corev1.ResourceList{
				corev1.ResourceRequestsCPU:    resource.MustParse("100m"),
				corev1.ResourceRequestsMemory: resource.MustParse("100Mi"),
			},
		},
	}
}
