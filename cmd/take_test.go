package cmd

import (
	"reflect"
	"testing"

	es "github.com/bebanjo/elastigo/lib"
)

func TestIndicesNames(t *testing.T) {
	indicesInfo := []es.CatIndexInfo{
		{
			Name: "index0",
		},
		{
			Name: "index1",
		},
		{
			Name: "index2",
		},
		{
			Name: "index3",
		},
		{
			Name: "index4",
		},
	}

	expectedIndicesNames := []string{
		"index0", "index1", "index2",
		"index3", "index4",
	}

	indicesNames := indicesNames(indicesInfo)
	if !reflect.DeepEqual(indicesNames, expectedIndicesNames) {
		t.Errorf("got %v, expected %v", indicesNames, expectedIndicesNames)
	}
}
