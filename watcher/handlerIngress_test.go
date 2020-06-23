package watcher

import (
	//	"encoding/json"

	//	t "ingress-ats/types"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func addBasicTest() {

}

func createExampleIngress() v1beta1.Ingress {
	exampleIngress := v1beta1.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "example-ingress",
			Namespace: "trafficserver-test",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: "cafe.example.com",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/coffee",
									Backend: v1beta1.IngressBackend{
										ServiceName: "coffee-svc",
										ServicePort: intstr.FromString("80"),
									},
								},
								{
									Path: "/tea",
									Backend: v1beta1.IngressBackend{
										ServiceName: "tea-svc",
										ServicePort: intstr.FromString("80"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return exampleIngress
}
