package watcher

import (
	"log"
	"reflect"
	"testing"

	ep "ingress-ats/endpoint"
	"ingress-ats/namespace"
	"ingress-ats/proxy"
	"ingress-ats/redis"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestBasic(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()

	igHandler.add(&exampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeys()

	var expectedKeys []interface{}
	expectedKeys = append(expectedKeys, "http://cafe.example.com/coffee")
	expectedKeys = append(expectedKeys, "http://cafe.example.com/tea")

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
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

func createExampleIgHandler() IgHandler {
	exampleEndpoint := createExampleEndpoint()
	igHandler := IgHandler{"ingresses", &exampleEndpoint}

	return igHandler
}

func createExampleEndpoint() ep.Endpoint {
	rClient, err := redis.Init()
	if err != nil {
		log.Panicln("Redis Error: ", err)
	}

	namespaceMap := make(map[string]bool)
	namespaceMap["trafficserver-test"] = true

	ignoreNamespaceMap := make(map[string]bool)

	nsManager := namespace.NsManager{
		NamespaceMap:       namespaceMap,
		IgnoreNamespaceMap: ignoreNamespaceMap,
	}

	exampleEndpoint := ep.Endpoint{
		RedisClient: rClient,
		ATSManager:  &proxy.ATSManager{Namespace: "default", IngressClass: ""},
		NsManager:   &nsManager,
	}

	return exampleEndpoint
}
