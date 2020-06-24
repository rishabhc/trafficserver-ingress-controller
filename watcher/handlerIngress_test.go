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

func TestAdd_ExampleIngress(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()

	igHandler.add(&exampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()

	expectedKeys := getExpectedKeysForAdd()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestAdd_ExampleIngressWithAnnotation(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngressWithAnnotation()

	igHandler.add(&exampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()

	expectedKeys := getExpectedKeysForAddWithAnnotation()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestUpdate_ModifyIngress(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()
	updatedExampleIngress := createExampleIngress()

	updatedExampleIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[1].Path = "/app2-modified"
	updatedExampleIngress.Spec.Rules[1].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName = "appsvc1-modified"
	updatedExampleIngress.Spec.Rules[1].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort = intstr.FromString("9090")

	igHandler.update(&exampleIngress, &updatedExampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()

	expectedKeys := getExpectedKeysForUpdate_ModifyIngress()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestUpdate_DeletePath(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()
	updatedExampleIngress := createExampleIngress()

	updatedExampleIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths = updatedExampleIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[:1]

	igHandler.update(&exampleIngress, &updatedExampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()
	expectedKeys := getExpectedKeysForUpdate_DeleteService()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDelete(t *testing.T) {
	igHandler := createExampleIgHandler()
	exampleIngress := createExampleIngress()

	igHandler.delete(&exampleIngress)

	returnedKeys := igHandler.Ep.RedisClient.GetDBOneKeyValues()

	expectedKeys := make(map[string][]string)

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}

}

func createExampleIngressWithAnnotation() v1beta1.Ingress {
	exampleIngress := createExampleIngress()

	exampleIngress.ObjectMeta.Annotations = make(map[string]string)
	exampleIngress.ObjectMeta.Annotations["ats.ingress.kubernetes.io/server-snippet"] = getExampleSnippet()
	exampleIngress.Spec.Rules = exampleIngress.Spec.Rules[1:]

	return exampleIngress
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

func getExpectedKeysForUpdate_ModifyIngress() map[string][]string {
	expectedKeys := getExpectedKeysForAdd()

	delete(expectedKeys, "http://test.media.com/app2")

	expectedKeys["http://test.media.com/app2-modified"] = []string{}
	expectedKeys["http://test.media.com/app2-modified"] = append(expectedKeys["http://test.media.com/app2"], "trafficserver-test:appsvc2:8080")

	expectedKeys["http://test.edge.com/app1"] = []string{}
	expectedKeys["http://test.edge.com/app1"] = append(expectedKeys["http://test.edge.com/app1"], "trafficserver-test:appsvc1-modified:9090")

	return expectedKeys
}

func getExpectedKeysForUpdate_DeleteService() map[string][]string {
	expectedKeys := getExpectedKeysForAdd()

	delete(expectedKeys, "http://test.media.com/app2")

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

func getExpectedKeysForAddWithAnnotation() map[string][]string {
	expectedKeys := getExpectedKeysForAdd()

	delete(expectedKeys, "http://test.media.com/app1")
	delete(expectedKeys, "http://test.media.com/app2")

	expectedKeys["http://test.edge.com/app1"] = append(expectedKeys["http://test.edge.com/app1"], "$trafficserver-test/example-ingress/")

	exampleSnippet := getExampleSnippet()

	expectedKeys["$trafficserver-test/example-ingress/"] = []string{}
	expectedKeys["$trafficserver-test/example-ingress/"] = append(expectedKeys["$trafficserver-test/example-ingress/"], exampleSnippet)

	return expectedKeys
}

func getExampleSnippet() string {
	return `ts.debug('Debug msg example')
	ts.error('Error msg example')
	-- ts.hook(TS_LUA_HOOK_SEND_RESPONSE_HDR, function()
	--   ts.client_response.header['Location'] = 'https://test.edge.com/app2'
	-- end)
	-- ts.http.skip_remapping_set(0)
	-- ts.http.set_resp(301, 'Redirect')
	ts.debug('Uncomment the above lines to redirect http request to https')`
}
