package cmd

import (
	"reflect"
	"testing"

	esclient "github.com/bebanjo/esnap/internal/es"
)

func TestIndicesNamesToDelete(t *testing.T) {
	indicesInfo := []esclient.Index{
		{Name: "index0"},
		{Name: "index1"},
		{Name: "index2"},
		{Name: "index3"},
		{Name: "index4"},
	}

	aliasesInfo := []esclient.Alias{
		{Name: "alias0", Index: "index3"},
		{Name: "alias1", Index: "index1"},
	}

	expectedIndicesNamesToDelete := []string{"index0", "index2", "index4"}

	indicesNamesToDelete := indicesNamesToDelete(indicesInfo, aliasesInfo)
	if !reflect.DeepEqual(indicesNamesToDelete, expectedIndicesNamesToDelete) {
		t.Errorf("got %v, expected %v", indicesNamesToDelete, expectedIndicesNamesToDelete)
	}
}
