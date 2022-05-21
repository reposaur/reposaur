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
		{"POST /happyPath", "POST", "/happyPath"},
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
		unevaluatedPath     string
		evaluatedPathParams []string
	}{
		{"/{potato}", []string{"potato"}},
		{"/{karma}/{koma}/yes", []string{"karma", "koma"}},
		{"/port/wine", []string{}},
	}

	data := map[string]any{
		"potato": 1,
		"karma":  "house",
		"koma":   "test",
	}

	for _, tc := range tests {
		testname := fmt.Sprintf("Getting evatluated params from %s", tc.unevaluatedPath)
		t.Run(testname, func(t *testing.T) {
			pathParams := parsePathParams(tc.unevaluatedPath)
			path := tc.unevaluatedPath
			evaluatedPathParams := tc.evaluatedPathParams

			url, _ := buildRequestUrl(path, data)

			fmt.Println(url)

			if !Equal(pathParams, evaluatedPathParams) {
				t.Errorf("got %v, want %v", pathParams, evaluatedPathParams)
			}
		})
	}
}

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

//Testing strings that have less than 2 arguments so error is expected
// func TestBuildRequestPathWithUnparsedRequestStringWithErrors2(t *testing.T) {
// 	var tests = []struct {
// 		unparsedReq string
// 	}{
// 		{"GET /happyPath/{potato}"},
// 		{"POST /happyPath"},
// 	}

// 	for _, tt := range tests {
// 		testname := fmt.Sprintf("Parsing %s", tt.unparsedReq)
// 		t.Run(testname, func(t *testing.T) {
// 			rp, _ := buildRequestPath(tt.unparsedReq)
// 			pathParams := parsePathParams(rp.path)

// 			var err error
// 			if err == nil {
// 				t.Errorf("got %v, want error", err)
// 			}
// 		})
// 	}
// }
