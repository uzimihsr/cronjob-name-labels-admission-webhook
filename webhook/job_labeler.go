package webhook

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	labelKey     string = "uzimihsr.github.io/cronjob-name"
	labelKeyPath string = "/metadata/labels/uzimihsr.github.io~1cronjob-name"
)

// Add a label {"uzimihsr.github.io/cronjob-name": "<CronJobName>"} to the Job object
func LabelJobOwnedByCronJob(ar v1.AdmissionReview) *v1.AdmissionResponse {
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if len(obj.ObjectMeta.OwnerReferences) == 0 {
		// the Job is not owned by anyone
		return &reviewResponse
	}

	for _, v := range obj.ObjectMeta.OwnerReferences {
		if v.Kind == "CronJob" {
			klog.Info(fmt.Sprintf("The Job(%s) is owned by CronJob(%s)", obj.ObjectMeta.Name, v.Name))
			labelValue, hasLabel := obj.ObjectMeta.Labels[labelKey]
			pt := v1.PatchTypeJSONPatch
			switch {
			case len(obj.ObjectMeta.Labels) == 0:
				addFirstLabelPatch := `[{ "op": "add", "path": "/metadata/labels", "value": {"%s": "%s"}}]`
				reviewResponse.Patch = []byte(fmt.Sprintf(addFirstLabelPatch, labelKey, v.Name))
				reviewResponse.PatchType = &pt
			case !hasLabel:
				addAdditionalLabelPatch := `[{ "op": "add", "path": "%s", "value": "%s" }]`
				reviewResponse.Patch = []byte(fmt.Sprintf(addAdditionalLabelPatch, labelKeyPath, v.Name))
				reviewResponse.PatchType = &pt
			case labelValue != v.Name:
				updateLabelPatch := `[{ "op": "replace", "path": "%s", "value": "%s" }]`
				reviewResponse.Patch = []byte(fmt.Sprintf(updateLabelPatch, labelKeyPath, v.Name))
				reviewResponse.PatchType = &pt
			default:
				// already set
			}
		}
	}

	return &reviewResponse
}
