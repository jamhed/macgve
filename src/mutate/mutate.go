package mutate

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GoVaultEnv struct {
	image     string
	vaultaddr string
	authpath  string
	container string
}

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
	if container, ok := pod.Annotations["govaultenv.io/container"]; !ok {
		gve.container = "*"
	} else {
		correct := false
		gve.container = container
		for _, c := range pod.Spec.Containers {
			if c.Name == gve.container {
				correct = true
			}
		}
		if !correct {
			log.Errorf("Can't find specified container to mutate: %v", gve.container)
			return nil
		}
	}
	if authpath, ok := pod.Annotations["govaultenv.io/authpath"]; !ok {
		return nil
	} else {
		gve.authpath = authpath
	}
	return gve
}

func Mutate(ar *v1beta1.AdmissionReview, vaultaddr, gveimage string) *v1beta1.AdmissionResponse {
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

	gve := mutable(pod)
	if gve == nil {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	origin := pod.DeepCopy()
	gve.vaultaddr = vaultaddr
	gve.image = gveimage
	pod = insertGve(pod, gve)

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
