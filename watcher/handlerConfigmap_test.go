package watcher

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAdd_BasicConfigMap(t *testing.T) {
	cmHandler := createExampleCMHandler()
	exampleConfigMap := createExampleConfigMap()

	cmHandler.Add(&exampleConfigMap)

	rEnabled, err := cmHandler.Ep.ATSManager.ConfigGet("proxy.config.output.logfile.rolling_enabled")

	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rEnabled, "1") {
		t.Errorf("returned \n%s,  but expected \n%s", rEnabled, "1")
	}

	rInterval, err := cmHandler.Ep.ATSManager.ConfigGet("proxy.config.output.logfile.rolling_interval_sec")

	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rInterval, "3000") {
		t.Errorf("returned \n%s,  but expected \n%s", rInterval, "3000")
	}

	threshold, err := cmHandler.Ep.ATSManager.ConfigGet("proxy.config.restart.active_client_threshold")

	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(threshold, "0") {
		t.Errorf("returned \n%s,  but expected \n%s", threshold, "0")
	}

}

func TestUpdate_BasicConfigMap(t *testing.T) {
	cmHandler := createExampleCMHandler()
	exampleConfigMap := createExampleConfigMap()
	exampleConfigMap.Data["proxy.config.output.logfile.rolling_interval_sec"] = "2000"

	cmHandler.update(&exampleConfigMap)

	rEnabled, err := cmHandler.Ep.ATSManager.ConfigGet("proxy.config.output.logfile.rolling_enabled")

	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rEnabled, "1") {
		t.Errorf("returned \n%s,  but expected \n%s", rEnabled, "1")
	}

	rInterval, err := cmHandler.Ep.ATSManager.ConfigGet("proxy.config.output.logfile.rolling_interval_sec")

	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rInterval, "2000") {
		t.Errorf("returned \n%s,  but expected \n%s", rInterval, "2000")
	}

	threshold, err := cmHandler.Ep.ATSManager.ConfigGet("proxy.config.restart.active_client_threshold")

	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(threshold, "0") {
		t.Errorf("returned \n%s,  but expected \n%s", threshold, "0")
	}

}

func createExampleConfigMap() v1.ConfigMap {
	exampleConfigMap := v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "testsvc",
			Namespace: "trafficserver-test-2",
		},
		Data: map[string]string{
			"proxy.config.output.logfile.rolling_enabled":      "1",
			"proxy.config.output.logfile.rolling_interval_sec": "3000",
			"proxy.config.restart.active_client_threshold":     "0",
		},
	}

	return exampleConfigMap
}

func createExampleCMHandler() CMHandler {
	exampleEndpoint := createExampleEndpoint()
	cmHandler := CMHandler{"configmap", &exampleEndpoint}

	return cmHandler
}
