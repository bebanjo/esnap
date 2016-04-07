package cmd

import (
	"reflect"
	"testing"

	es "github.com/bebanjo/esnap/vendor/src/github.com/mattbaird/elastigo/lib"
)

func TestIndicesNames(t *testing.T) {
	indicesInfo := []es.CatIndexInfo{
		es.CatIndexInfo{
			Name: "index0",
		},
		es.CatIndexInfo{
			Name: "index1",
		},
		es.CatIndexInfo{
			Name: "index2",
		},
		es.CatIndexInfo{
			Name: "index3",
		},
		es.CatIndexInfo{
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
