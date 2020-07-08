package watcher

import (
	"ingress-ats/util"
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	fake "k8s.io/client-go/kubernetes/fake"
	fakecontroller "k8s.io/client-go/tools/cache/testing"
	framework "k8s.io/client-go/tools/cache/testing"
)

func TestAllNamespacesWatchFor_Add(t *testing.T) {
	w, fc := getTestWatcher()

	epHandler := EpHandler{"endpoints", w.Ep}
	err := w.allNamespacesWatchFor(&epHandler, w.Cs.CoreV1().RESTClient(),
		fields.Everything(), &v1.Endpoints{}, 0, fc)

	if err != nil {
		t.Error(err)
	}

	fc.Add(&v1.Endpoints{
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
	})
	time.Sleep(100 * time.Millisecond)

	returnedKeys := w.Ep.RedisClient.GetDefaultDBKeyValues()
	expectedKeys := getExpectedKeysForEndpointAdd()
	expectedKeys["trafficserver-test-2:testsvc:8080"] = util.ReverseSlice(expectedKeys["trafficserver-test-2:testsvc:8080"])

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestAllNamespacesWatchFor_Update(t *testing.T) {
	w, fc := getTestWatcher()

	epHandler := EpHandler{"endpoints", w.Ep}
	err := w.allNamespacesWatchFor(&epHandler, w.Cs.CoreV1().RESTClient(),
		fields.Everything(), &v1.Endpoints{}, 0, fc)

	if err != nil {
		t.Error(err)
	}

	fc.Add(&v1.Endpoints{
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
	})
	time.Sleep(100 * time.Millisecond)

	fc.Modify(&v1.Endpoints{
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
						IP: "10.10.3.3",
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
	})
	time.Sleep(100 * time.Millisecond)

	returnedKeys := w.Ep.RedisClient.GetDefaultDBKeyValues()
	expectedKeys := getExpectedKeysForEndpointAdd()
	expectedKeys["trafficserver-test-2:testsvc:8080"][1] = "10.10.3.3#8080#http"

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestAllNamespacesWatchFor_Delete(t *testing.T) {
	w, fc := getTestWatcher()

	epHandler := EpHandler{"endpoints", w.Ep}
	err := w.allNamespacesWatchFor(&epHandler, w.Cs.CoreV1().RESTClient(),
		fields.Everything(), &v1.Endpoints{}, 0, fc)

	if err != nil {
		t.Error(err)
	}

	fc.Add(&v1.Endpoints{
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
	})
	time.Sleep(100 * time.Millisecond)

	fc.Delete(&v1.Endpoints{
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
						IP: "10.10.3.3",
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
	})
	time.Sleep(100 * time.Millisecond)

	returnedKeys := w.Ep.RedisClient.GetDefaultDBKeyValues()
	expectedKeys := make(map[string][]string)

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func getTestWatcher() (Watcher, *framework.FakeControllerSource) {
	clientset := fake.NewSimpleClientset()
	fc := fakecontroller.NewFakeControllerSource()

	exampleEndpoint := createExampleEndpoint()
	stopChan := make(chan struct{})

	ingressWatcher := Watcher{
		Cs:           clientset,
		ATSNamespace: "default",
		Ep:           &exampleEndpoint,
		StopChan:     stopChan,
	}

	return ingressWatcher, fc
}
