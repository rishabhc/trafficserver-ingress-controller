package watcher

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAdd(t *testing.T) {
	epHandler := createExampleEpHandler()
	exampleV1Endpoint := createExampleV1Endpoint()

	epHandler.add(&exampleV1Endpoint)

	returnedKeys := epHandler.Ep.RedisClient.GetDefaultDBKeyValues()

	expectedKeys := getExpectedKeysForEndpointAdd()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func createExampleV1Endpoint() v1.Endpoints {
	exampleEndpoint := v1.Endpoints{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "testsvc",
			Namespace: "trafficserver-test-2",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "10.10.1.1",
					},
					{
						IP: "10.10.2.2",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "main",
						Port:     8080,
						Protocol: "TCP",
					},
				},
			},
		},
	}

	return exampleEndpoint
}

func createExampleEpHandler() EpHandler {
	exampleEndpoint := createExampleEndpoint()
	epHandler := EpHandler{"endpoints", &exampleEndpoint}

	return epHandler
}

func getExpectedKeysForEndpointAdd() map[string][]string {
	expectedKeys := make(map[string][]string)
	expectedKeys["trafficserver-test-2:testsvc:8080"] = []string{}

	expectedKeys["trafficserver-test-2:testsvc:8080"] = append(expectedKeys["trafficserver-test-2:testsvc:8080"], "10.10.1.1#8080#http")
	expectedKeys["trafficserver-test-2:testsvc:8080"] = append(expectedKeys["trafficserver-test-2:testsvc:8080"], "10.10.2.2#8080#http")

	return expectedKeys
}
