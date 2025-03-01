package assets

import (
	"fmt"
	. "github.com/onsi/gomega"
	hyperv1 "github.com/openshift/hypershift/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHyperShiftOperatorDeployment_Build(t *testing.T) {
	testNamespace := "hypershift"
	testOperatorImage := "myimage"
	tests := map[string]struct {
		inputBuildParameters HyperShiftOperatorDeployment
		expectedVolumeMounts []corev1.VolumeMount
		expectedVolumes      []corev1.Volume
		expectedArgs         []string
	}{
		"empty oidc paramaters result in no volume mounts": {
			inputBuildParameters: HyperShiftOperatorDeployment{
				Namespace: &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNamespace,
					},
				},
				OperatorImage: testOperatorImage,
				ServiceAccount: &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hypershift",
					},
				},
				Replicas:        3,
				PrivatePlatform: string(hyperv1.NonePlatform),
			},
			expectedVolumeMounts: nil,
			expectedVolumes:      nil,
			expectedArgs: []string{
				"run",
				"--namespace=$(MY_NAMESPACE)",
				"--deployment-name=operator",
				"--metrics-addr=:9000",
				fmt.Sprintf("--enable-ocp-cluster-monitoring=%t", false),
				fmt.Sprintf("--enable-ci-debug-output=%t", false),
				fmt.Sprintf("--private-platform=%s", string(hyperv1.NonePlatform)),
			},
		},
		"specify oidc parameters result in appropriate volumes and volumeMounts": {
			inputBuildParameters: HyperShiftOperatorDeployment{
				Namespace: &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNamespace,
					},
				},
				OperatorImage: testOperatorImage,
				ServiceAccount: &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hypershift",
					},
				},
				Replicas:         3,
				PrivatePlatform:  string(hyperv1.AWSPlatform),
				OIDCBucketRegion: "us-east-1",
				OIDCStorageProviderS3Secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "oidc-s3-secret",
					},
				},
				OIDCBucketName:                 "oidc-bucket",
				OIDCStorageProviderS3SecretKey: "mykey",
			},
			expectedArgs: []string{
				"run",
				"--namespace=$(MY_NAMESPACE)",
				"--deployment-name=operator",
				"--metrics-addr=:9000",
				fmt.Sprintf("--enable-ocp-cluster-monitoring=%t", false),
				fmt.Sprintf("--enable-ci-debug-output=%t", false),
				fmt.Sprintf("--private-platform=%s", string(hyperv1.AWSPlatform)),
				"--oidc-storage-provider-s3-bucket-name=" + "oidc-bucket",
				"--oidc-storage-provider-s3-region=" + "us-east-1",
				"--oidc-storage-provider-s3-credentials=/etc/oidc-storage-provider-s3-creds/" + "mykey",
			},
			expectedVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "oidc-storage-provider-s3-creds",
					MountPath: "/etc/oidc-storage-provider-s3-creds",
				},
				{
					Name:      "credentials",
					MountPath: "/etc/provider",
				},
				{
					Name:      "token",
					MountPath: "/var/run/secrets/openshift/serviceaccount",
				},
			},
			expectedVolumes: []corev1.Volume{
				{
					Name: "oidc-storage-provider-s3-creds",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "oidc-s3-secret",
						},
					},
				},
				{
					Name: "credentials",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: awsCredsSecretName,
						},
					},
				},
				{
					Name: "token",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
										Audience: "openshift",
										Path:     "token",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			deployment := test.inputBuildParameters.Build()
			g.Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(BeEquivalentTo(test.expectedArgs))
			g.Expect(deployment.Spec.Template.Spec.Volumes).To(BeEquivalentTo(test.expectedVolumes))
			g.Expect(deployment.Spec.Template.Spec.Containers[0].VolumeMounts).To(BeEquivalentTo(test.expectedVolumeMounts))
		})
	}
}
