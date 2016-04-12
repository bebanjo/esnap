package cmd

import (
	"reflect"
	"testing"

	es "github.com/bebanjo/elastigo/lib"
)

func TestIndicesNamesToDelete(t *testing.T) {
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

	aliasesInfo := []es.CatAliasInfo{
		es.CatAliasInfo{
			Name:  "alias0",
			Index: "index3",
		},
		es.CatAliasInfo{
			Name:  "alias1",
			Index: "index1",
		},
	}

	expectedIndicesNamesToDelete := []string{
		"index0", "index2", "index4",
	}

	indicesNamesToDelete := indicesNamesToDelete(indicesInfo, aliasesInfo)
	if !reflect.DeepEqual(indicesNamesToDelete, expectedIndicesNamesToDelete) {
		t.Errorf("got %v, expected %v", indicesNamesToDelete, expectedIndicesNamesToDelete)
	}
}
