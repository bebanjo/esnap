package cmd

import (
	"reflect"
	"testing"

	es "github.com/bebanjo/esnap/vendor/src/github.com/mattbaird/elastigo/lib"
)

func TestIndicesToRemove(t *testing.T) {
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

	expectedIndicesToRemove := []string{
		"index0", "index2", "index4",
	}

	indicesToRemove := indicesToRemove(indicesInfo, aliasesInfo)
	if !reflect.DeepEqual(indicesToRemove, expectedIndicesToRemove) {
		t.Errorf("got %v, expected %v", indicesToRemove, expectedIndicesToRemove)
	}

}
