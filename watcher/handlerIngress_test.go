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

func TestBasicAdd(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()

	igHandler.add(&exampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()

	expectedKeys := getExpectedKeysForAdd()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestBasicUpdate(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()
	updatedExampleIngress := createExampleIngress()

	updatedExampleIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[1].Path = "/app2-modified"

	igHandler.update(&exampleIngress, &updatedExampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()

	expectedKeys := getExpectedKeysForUpdate()

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
					Host: "test.media.com",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/app1",
									Backend: v1beta1.IngressBackend{
										ServiceName: "appsvc1",
										ServicePort: intstr.FromString("8080"),
									},
								},
								{
									Path: "/app2",
									Backend: v1beta1.IngressBackend{
										ServiceName: "appsvc2",
										ServicePort: intstr.FromString("8080"),
									},
								},
							},
						},
					},
				},
				{
					Host: "test.edge.com",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/app1",
									Backend: v1beta1.IngressBackend{
										ServiceName: "appsvc1",
										ServicePort: intstr.FromString("8080"),
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

func getExpectedKeysForUpdate() map[string][]string {
	expectedKeys := getExpectedKeysForAdd()

	delete(expectedKeys, "http://test.media.com/app2")

	expectedKeys["http://test.media.com/app2-modified"] = []string{}
	expectedKeys["http://test.media.com/app2-modified"] = append(expectedKeys["http://test.media.com/app2"], "trafficserver-test:appsvc2:8080")

	return expectedKeys
}

func getExpectedKeysForAdd() map[string][]string {
	expectedKeys := make(map[string][]string)
	expectedKeys["http://test.edge.com/app1"] = []string{}
	expectedKeys["http://test.media.com/app1"] = []string{}
	expectedKeys["http://test.media.com/app2"] = []string{}

	expectedKeys["http://test.edge.com/app1"] = append(expectedKeys["http://test.edge.com/app1"], "trafficserver-test:appsvc1:8080")
	expectedKeys["http://test.media.com/app2"] = append(expectedKeys["http://test.media.com/app2"], "trafficserver-test:appsvc2:8080")
	expectedKeys["http://test.media.com/app1"] = append(expectedKeys["http://test.media.com/app1"], "trafficserver-test:appsvc1:8080")

	return expectedKeys
}
