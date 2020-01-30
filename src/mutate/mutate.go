package mutate

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makePatch(obj, copyObj *corev1.Pod) ([]jsonpatch.Operation, error) {
	origJSON, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	mutatedJSON, err := json.Marshal(copyObj)
	if err != nil {
		return nil, err
	}
	patch, err := jsonpatch.CreatePatch(origJSON, mutatedJSON)
	if err != nil {
		return nil, err
	}
	return patch, nil
}

func maybeMutate(containers []corev1.Container) {
	for i, container := range containers {
		insertVar := true
		for key, value := range container.Resources.Limits {
			if key == "nvidia.com/gpu" && value.Value() > 0 {
				insertVar = false
			}
		}
		if insertVar {
			if container.Env == nil {
				container.Env = []corev1.EnvVar{}
			}
			container.Env = append(container.Env, corev1.EnvVar{Name: "NVIDIA_VISIBLE_DEVICES", Value: "none"})
		}
		log.Debugf("Inserting environment variable to container:%s, %s", container.Name, container.Env)
		containers[i] = container
	}
}

func Mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	var pod *corev1.Pod
	req := ar.Request
	log.Infof("Mutate kind:%s namespace:%s, name:%s, op:%s, userinfo:%s, uid:%s", req.Kind, req.Namespace, req.Name, req.Operation, req.UserInfo, req.UID)

	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		log.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}

	origin := pod.DeepCopy()
	maybeMutate(pod.Spec.Containers)

	patch, err := makePatch(origin, pod)
	if err != nil {
		log.Errorf("Could not make patch, error: %v", err)
		return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}
	bytes, err := json.Marshal(patch)
	if err != nil {
		log.Errorf("Could not marshal patch object: %v", err)
		return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}
	log.Debugf("Mutated response: %s", string(bytes))

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   bytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}
