package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	esclient "github.com/bebanjo/esnap/internal/es"
)

func TestPartitionRestoredIndices(t *testing.T) {
	// Naming scheme mandated by MGXSAAS-5955: restored indices are named
	// <destination>_<alias>_<restoreDate>, with no snapshot ID appended.
	const (
		currentAliasedIndex = "staging_movida_schedulings_20250122194121"
		restoredIndex       = "staging_movida_schedulings_20260721063621"
		staleIndex          = "staging_movida_schedulings_20260101000000"
		suffix              = "20260721063621"
	)

	indicesNames := []string{currentAliasedIndex, restoredIndex, staleIndex}

	toDelete, toAlias := partitionRestoredIndices(indicesNames, currentAliasedIndex, suffix)

	wantToDelete := []string{currentAliasedIndex, staleIndex}
	wantToAlias := []string{restoredIndex}

	if !reflect.DeepEqual(toDelete, wantToDelete) {
		t.Fatalf("toDelete = %v, want %v", toDelete, wantToDelete)
	}
	if !reflect.DeepEqual(toAlias, wantToAlias) {
		t.Fatalf("toAlias = %v, want %v", toAlias, wantToAlias)
	}
}

func newRestoreCaptureClient(t *testing.T) (*esclient.Client, *string) {
	t.Helper()
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"accepted":true}`))
	}))
	t.Cleanup(server.Close)

	client, err := esclient.NewClient(esclient.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, &capturedBody
}

func TestRestoreRenameReplacementHasNoDoubleTimestamp(t *testing.T) {
	const (
		origin      = "production"
		destination = "staging"
		snapshotID  = "20260721040001"
		date        = "20260721063621"
	)

	client, body := newRestoreCaptureClient(t)

	if err := restore(client, origin, destination, snapshotID, date); err != nil {
		t.Fatalf("restore() error = %v", err)
	}

	want := `"rename_replacement":"staging_$1_20260721063621"`
	if !strings.Contains(*body, want) {
		t.Fatalf("rename_replacement missing or malformed, got body: %s", *body)
	}
	if strings.Contains(*body, snapshotID) {
		t.Fatalf("rename_replacement must not contain snapshot ID, got body: %s", *body)
	}
}

func TestFreshRestoreRenameReplacementHasNoDoubleTimestamp(t *testing.T) {
	const (
		origin      = "production"
		destination = "staging"
		snapshotID  = "20260721040001"
		date        = "20260721063621"
	)

	client, body := newRestoreCaptureClient(t)

	if err := freshRestore(client, origin, destination, snapshotID, date); err != nil {
		t.Fatalf("freshRestore() error = %v", err)
	}

	want := `"rename_replacement":"staging_$1_20260721063621"`
	if !strings.Contains(*body, want) {
		t.Fatalf("rename_replacement missing or malformed, got body: %s", *body)
	}
	if strings.Contains(*body, snapshotID) {
		t.Fatalf("rename_replacement must not contain snapshot ID, got body: %s", *body)
	}
}
