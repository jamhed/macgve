package mutate

import corev1 "k8s.io/api/core/v1"

func mutateContainer(c corev1.Container, gve *GoVaultEnv) corev1.Container {
	args := append(c.Command, c.Args...)
	c.Command = []string{"/vault/govaultenv", "-addr=" + gve.vaultaddr, "-kubeauth=" + gve.authpath, "-stripname=true"}
	c.Args = args
	c.VolumeMounts = append(c.VolumeMounts, []corev1.VolumeMount{
		{Name: "govaultenv", MountPath: "/vault"},
		{Name: "tls-certs", MountPath: "/etc/ssl"},
	}...)
	return c
}

func insertGve(pod *corev1.Pod, gve *GoVaultEnv) *corev1.Pod {
	for i, c := range pod.Spec.Containers {
		if c.Name == gve.container {
			pod.Spec.Containers[i] = mutateContainer(c, gve)
		}
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, makeInitContainer(gve.image))
	pod.Spec.Volumes = append(pod.Spec.Volumes, makeVolumes()...)
	return pod
}

func makeVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name:         "govaultenv",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}},
		},
		{
			Name:         "tls-certs",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}},
		},
	}
}

func makeInitContainer(gveImage string) corev1.Container {
	return corev1.Container{
		Name:            "govaultenv-init",
		Image:           gveImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"sh", "-c", "cp /govaultenv /vault/ && cp /ca-certificates.crt /etc/ssl/cert.pem"},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "govaultenv", MountPath: "/vault"},
			{Name: "tls-certs", MountPath: "/etc/ssl"},
		},
	}
}
