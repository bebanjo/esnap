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

func TestAliasedIndicesNames(t *testing.T) {
	aliasesInfo := []es.CatAliasInfo{
		{
			Name:  "index0",
			Index: "index0_123",
		},
		{
			Name:  "index1",
			Index: "index1_123",
		},
		{
			Name:  "index2",
			Index: "index2_123",
		},
		{
			Name:  "index3",
			Index: "index3_123",
		},
		{
			Name:  "index4",
			Index: "index4_123",
		},
	}

	indicesNames := []string{
		"index0_123", "index0_456",
		"index1_123", "index1_456",
		"index2_123", "index2_456",
		"index3_123", "index3_456",
		"index4_123", "index4_456",
	}

	expectedAliasedIndices := []string{
		"index0_123", "index1_123", "index2_123",
		"index3_123", "index4_123",
	}

	aliasedIndices := aliasedIndicesNames(aliasesInfo, indicesNames)
	if !reflect.DeepEqual(aliasedIndices, expectedAliasedIndices) {
		t.Errorf("got %v, expected %v", aliasedIndices, expectedAliasedIndices)
	}
}
