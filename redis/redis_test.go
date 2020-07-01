package redis

import (
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	_, err := Init()
	if err != nil {
		t.Error(err)
	}
}

func TestDefaultDBSAdd(t *testing.T) {
	rClient, _ := Init()

	rClient.DefaultDBSAdd("test-key", "test-val")
	returnedKeys := rClient.GetDefaultDBKeyValues()
	expectedKeys := getExpectedKeysForAdd()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDefaultDBDel(t *testing.T) {
	rClient, _ := Init()

	rClient.DefaultDBSAdd("test-key", "test-val")
	rClient.DefaultDBSAdd("test-key-2", "test-val-2")
	rClient.DefaultDBDel("test-key")

	returnedKeys := rClient.GetDefaultDBKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	delete(expectedKeys, "test-key")
	expectedKeys["test-key-2"] = make([]string, 1)
	expectedKeys["test-key-2"][0] = "test-val-2"

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDefaultDBSUnionStore(t *testing.T) {
	rClient, _ := Init()

	rClient.DefaultDBSAdd("test-key", "test-val")
	rClient.DefaultDBSAdd("test-key-2", "test-val-2")
	rClient.DefaultDBSUnionStore("test-key", "test-key-2")

	returnedKeys := rClient.GetDefaultDBKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func getExpectedKeysForAdd() map[string][]string {
	expectedKeys := make(map[string][]string)
	expectedKeys["test-key"] = make([]string, 1)
	expectedKeys["test-key"][0] = "test-val"
	return expectedKeys
}
