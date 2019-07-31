package agent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestGCPMetaDataGetSuffix(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		switch path := r.URL.EscapedPath(); path {
		case "/computeMetadata/v1/value":
			fmt.Fprintf(w, "Wiggy Wiggy")
		case "/computeMetadata/v1/nested/paths/work":
			fmt.Fprintf(w, "Velociraptors are terrifying")
		case "/computeMetadata/v1/work":
			fmt.Fprintf(w, "This is a silly path")
		case "/computeMetadata/v1/messed-up":
			fmt.Fprintf(w, "Sourdough is the greatest")
		default:
			t.Fatalf("Error %q", path)
		}
	}))
	defer ts.Close()

	url, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Error %q", err)
	}

	old := os.Getenv("GCE_METADATA_HOST")
	defer os.Setenv("GCE_METADATA_HOST", old)
	os.Setenv("GCE_METADATA_HOST", url.Host)

	values, err := GCPMetaData{}.GetSuffixes([]string{
		"simple=value",
		"nested=nested/paths/work",
		"silly_path=silly/paths/../../work",
		"weird_path=///messed-up",
		"weird key=value",
	})

	if err != nil {
		t.Fatalf("Error %q", err)
	}

	assert.Equal(t, values, map[string]string{
		"simple":     "Wiggy Wiggy",
		"nested":     "Velociraptors are terrifying",
		"silly_path": "This is a silly path",
		"weird_path": "Sourdough is the greatest",
		"weird key":  "Wiggy Wiggy",
	})
}
