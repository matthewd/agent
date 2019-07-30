package agent

import (
	"fmt"
	"github.com/buildkite/agent/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchTags(t *testing.T) {
	// for _, tc := range []struct {
	// 	Destination, Path string
	// }{
	// 	{"gs://my-bucket-name/foo/bar", "foo/bar"},
	// 	{"gs://starts-with-an-s/and-this-is-its/folder", "and-this-is-its/folder"},
	// } {
	// 	_, path := FetchTags(tc.Destination)
	// 	if path != tc.Path {
	// 		t.Fatalf("Expected %q, got %q", tc.Path, path)
	// 	}
	// }

	tags := FetchTags(logger.Discard, FetchTagsConfig{
		Tags:                []string{"foo=bar", "hello=world"},
		TagsFromEC2MetaData: []string{"key=foo/bar"},
		TagsFromGCPMetaData: []string{"key=foo/bar"},
	})

	fmt.Printf("derp = %#v \n", tags)

	assert.Equal(t, tags, []string{"foo=bar", "hello=world"})
}

// func TestFetchTagsFromHost(t *testing.T) {
// 	tags := FetchTags(logger.Discard, FetchTagsConfig{
// 		Tags:         []string{"nope=not-there"},
// 		TagsFromHost: true,
// 	})

// 	assert.Equal(t, tags, []string{
// 		"hostname=corretto.local",
// 		"os=darwin",
// 		"machine-id=15053ee12d72986fb76bd9783b6859390fc1cab264da6023e697a6526b38acf",
// 	})
// }
