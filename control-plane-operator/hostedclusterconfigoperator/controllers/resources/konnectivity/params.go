package konnectivity

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	hyperv1 "github.com/openshift/hypershift/api/v1alpha1"
	"github.com/openshift/hypershift/support/config"
)

const (
	systemNodeCriticalPriorityClass = "system-node-critical"
)

type KonnectivityParams struct {
	Image           string
	ExternalAddress string
	ExternalPort    int32
	config.DeploymentConfig
}

func NewKonnectivityParams(hcp *hyperv1.HostedControlPlane, images map[string]string, externalAddress string, externalPort int32) *KonnectivityParams {
	p := &KonnectivityParams{
		Image:           images["konnectivity-agent"],
		ExternalAddress: externalAddress,
		ExternalPort:    externalPort,
	}

	p.DeploymentConfig.Resources = config.ResourcesSpec{
		konnectivityAgentContainer().Name: {
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("50Mi"),
				corev1.ResourceCPU:    resource.MustParse("40m"),
			},
		},
	}
	p.DeploymentConfig.Scheduling = config.Scheduling{
		PriorityClass: systemNodeCriticalPriorityClass,
	}
	p.DeploymentConfig.LivenessProbes = config.LivenessProbes{
		konnectivityAgentContainer().Name: {
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Scheme: corev1.URISchemeHTTP,
					Port:   intstr.FromInt(int(healthPort)),
					Path:   "healthz",
				},
			},
			InitialDelaySeconds: 120,
			TimeoutSeconds:      30,
			PeriodSeconds:       60,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		},
	}
	if _, ok := hcp.Annotations[hyperv1.KonnectivityAgentImageAnnotation]; ok {
		p.Image = hcp.Annotations[hyperv1.KonnectivityAgentImageAnnotation]
	}
	p.DeploymentConfig.SetReleaseImageAnnotation(hcp.Spec.ReleaseImage)
	p.DeploymentConfig.SetRestartAnnotation(hcp.ObjectMeta)
	return p
}
