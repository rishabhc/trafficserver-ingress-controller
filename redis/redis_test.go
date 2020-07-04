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

func TestFlush(t *testing.T) {
	rClient, _ := Init()
	rClient.DefaultDB.SAdd("test-key", "test-val")
	rClient.DefaultDB.SAdd("test-key", "test-val-2")

	err := rClient.Flush()
	if err != nil {
		t.Error(err)
	}

	returnedKeys := rClient.GetDefaultDBKeyValues()
	expectedKeys := make(map[string][]string)
	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestGetDefaultDBKeyValues(t *testing.T) {
	rClient, _ := Init()

	rClient.DefaultDB.SAdd("test-key", "test-val")
	rClient.DefaultDB.SAdd("test-key", "test-val-2")
	rClient.DefaultDB.SAdd("test-key-2", "test-val")

	returnedKeys := rClient.GetDefaultDBKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	expectedKeys["test-key"] = append([]string{"test-val-2"}, expectedKeys["test-key"]...)
	expectedKeys["test-key-2"] = make([]string, 1)
	expectedKeys["test-key-2"][0] = "test-val"

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestGetDBOneKeyValues(t *testing.T) {
	rClient, _ := Init()

	rClient.DBOne.SAdd("test-key", "test-val")
	rClient.DBOne.SAdd("test-key", "test-val-2")
	rClient.DBOne.SAdd("test-key-2", "test-val")

	returnedKeys := rClient.GetDBOneKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	expectedKeys["test-key"] = append([]string{"test-val-2"}, expectedKeys["test-key"]...)
	expectedKeys["test-key-2"] = make([]string, 1)
	expectedKeys["test-key-2"][0] = "test-val"

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
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
	expectedKeys["test-key"][0] = "test-val-2"
	expectedKeys["test-key-2"] = make([]string, 1)
	expectedKeys["test-key-2"][0] = "test-val-2"

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDBOneSAdd(t *testing.T) {
	rClient, _ := Init()

	rClient.DBOneSAdd("test-key", "test-val")
	returnedKeys := rClient.GetDBOneKeyValues()
	expectedKeys := getExpectedKeysForAdd()

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDBOneSRem(t *testing.T) {
	rClient, _ := Init()

	rClient.DBOneSAdd("test-key", "test-val")
	rClient.DBOneSAdd("test-key", "test-val-2")
	rClient.DBOneSAdd("test-key", "test-val-3")
	rClient.DBOneSRem("test-key", "test-val-2")
	returnedKeys := rClient.GetDBOneKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	expectedKeys["test-key"] = append([]string{"test-val-3"}, expectedKeys["test-key"]...)

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDBOneDel(t *testing.T) {
	rClient, _ := Init()

	rClient.DBOneSAdd("test-key", "test-val")
	rClient.DBOneSAdd("test-key-2", "test-val-2")
	rClient.DBOneDel("test-key")

	returnedKeys := rClient.GetDBOneKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	delete(expectedKeys, "test-key")
	expectedKeys["test-key-2"] = make([]string, 1)
	expectedKeys["test-key-2"][0] = "test-val-2"

	if !reflect.DeepEqual(returnedKeys, expectedKeys) {
		t.Errorf("returned \n%v,  but expected \n%v", returnedKeys, expectedKeys)
	}
}

func TestDBOneSUnionStore(t *testing.T) {
	rClient, _ := Init()

	rClient.DBOneSAdd("test-key", "test-val")
	rClient.DBOneSAdd("test-key-2", "test-val-2")
	rClient.DBOneSUnionStore("test-key", "test-key-2")

	returnedKeys := rClient.GetDBOneKeyValues()
	expectedKeys := getExpectedKeysForAdd()
	expectedKeys["test-key"][0] = "test-val-2"
	expectedKeys["test-key-2"] = make([]string, 1)
	expectedKeys["test-key-2"][0] = "test-val-2"

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
