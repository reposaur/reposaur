package builtin

import (
	"fmt"
	"testing"
)

// Testing valid strings for unsplit request path
func TestBuildRequestPathWithUnsplitRequestString(t *testing.T) {
	var tests = []struct {
		unparsedReq string
		wantMethod  string
		wantPath    string
	}{
		{"GET /happyPath", "GET", "/happyPath"},
		{"GeT /happyPath", "GeT", "/happyPath"},
		{"GET /happy/Path", "GET", "/happy/Path"},
		{"get /happy/Path", "GET", "/happy/Path"},
	}

	for _, tc := range tests {
		testname := fmt.Sprintf("%s, %v + %v ", tc.unparsedReq, tc.wantMethod, tc.wantPath)
		t.Run(testname, func(t *testing.T) {
			method, path, _ := splitPath(tc.unparsedReq)

			wantMethod := tc.wantMethod
			wantPath := tc.wantPath
			if method != wantMethod && path != wantPath {
				t.Errorf("got %v and %v, want %v and %v", method, path, wantMethod, wantPath)
			}
		})
	}
}

//Testing strings that have less than 2 arguments so error is expected
func TestBuildRequestPathWithUnparsedRequestStringWithErrors(t *testing.T) {
	var tests = []struct {
		unevaluatedPath string
		parsedUrl       string
	}{
		{"/{potato}", "/1?karma=house&koma=test"},
		{"/{karma}/{koma}/yes", "/house/test/yes?potato=1"},
		{"/port/wine", "/port/wine?karma=house&koma=test&potato=1"},
	}

	for _, tc := range tests {

		data := map[string]any{
			"potato": 1,
			"karma":  "house",
			"koma":   "test",
		}
		testname := fmt.Sprintf("Evaluating params from %s", tc.unevaluatedPath)
		t.Run(testname, func(t *testing.T) {
			url, err := buildRequestUrl(tc.unevaluatedPath, data)

			if err != nil {
				t.Errorf("got %v, want %v", err, tc.parsedUrl)
			}
			if tc.parsedUrl != url {
				t.Errorf("got %v, want %v", url, tc.parsedUrl)
			}
		})
	}
}
