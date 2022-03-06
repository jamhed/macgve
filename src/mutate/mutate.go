package mutate

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	v1 "k8s.io/api/admission/v1"
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

func mutable(pod *corev1.Pod) *GoVaultEnv {
	gve := &GoVaultEnv{}
	if pod.Annotations == nil {
		return nil
	}
	if containers, ok := pod.Annotations["govaultenv.io/containers"]; ok {
		correct := false
		gve.SetContainers(containers)
		for _, c := range pod.Spec.Containers {
			if gve.IsIn(c.Name) {
				correct = true
			}
		}
		if !correct {
			log.Errorf("Can't find specified container to mutate: %v", gve.containers)
			return nil
		}
		log.Debugf("Containers to mutate:%s", gve.containers)
	}
	if authpath, ok := pod.Annotations["govaultenv.io/authpath"]; ok {
		gve.authpath = strings.Join([]string{pod.Spec.ServiceAccountName, authpath}, "@")
		return gve
	}
	if authpath, err := getFromNamespace(pod.Namespace); err == nil {
		if len(authpath) == 0 {
			return nil
		}
		gve.authpath = strings.Join([]string{pod.Spec.ServiceAccountName, authpath}, "@")
		return gve
	} else {
		log.Errorf("Error getting annotation from namespace: %v", err)
		return nil
	}
}

func Mutate(ar *v1.AdmissionReview, vaultaddr, gveimage string) *v1.AdmissionResponse {
	var pod *corev1.Pod
	req := ar.Request
	log.Infof("Mutate kind:%s namespace:%s, name:%s, op:%s, userinfo:%s, uid:%s", req.Kind, req.Namespace, req.Name, req.Operation, req.UserInfo, req.UID)

	if req.Kind.Kind != "Pod" {
		return &v1.AdmissionResponse{Allowed: true}
	}

	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		log.Errorf("Could not unmarshal raw object: %v", err)
		return &v1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}

	gve := mutable(pod)
	if gve == nil {
		return &v1.AdmissionResponse{Allowed: true}
	}

	origin := pod.DeepCopy()
	gve.vaultaddr = vaultaddr
	gve.image = gveimage
	pod = insertGve(pod, gve)

	patch, err := makePatch(origin, pod)
	if err != nil {
		log.Errorf("Could not make patch, error: %v", err)
		return &v1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}
	bytes, err := json.Marshal(patch)
	if err != nil {
		log.Errorf("Could not marshal patch object: %v", err)
		return &v1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}
	log.Debugf("Mutated response: %s", string(bytes))

	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   bytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}
