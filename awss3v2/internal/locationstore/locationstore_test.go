// Copyright 2021 The s3iot authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package locationstore

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	mock_s3iface "github.com/at-wat/s3iot/awss3v2/internal/moq/s3iface"
)

func TestLocationStore(t *testing.T) {
	errDummy := errors.New("error")

	testCases := map[string]struct {
		status     string
		statusCode int
		err        error
		expected   *http.Response
	}{
		"Success": {
			status:     "200 OK",
			statusCode: 200,
			expected: &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Request: &http.Request{
					URL: &url.URL{Scheme: "s3", Host: "url"},
				},
			},
		},
		"Fail": {
			status:     "404 NotFound",
			statusCode: 404,
			expected: &http.Response{
				Status:     "404 NotFound",
				StatusCode: 404,
				Request: &http.Request{
					URL: &url.URL{Scheme: "s3", Host: "url"},
				},
			},
			err: errDummy,
		},
	}
	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			s := LocationStore{
				HTTPClient: &mock_s3iface.MockHTTPClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{
							Status:     tt.status,
							StatusCode: tt.statusCode,
							Request:    request,
						}, tt.err
					},
				},
			}
			resp, err := s.Do(&http.Request{
				URL: &url.URL{Scheme: "s3", Host: "url"},
			})

			if !reflect.DeepEqual(tt.expected, resp) {
				t.Errorf("Expected response: %+v, got: %+v", tt.expected, resp)
			}
			if err != tt.err {
				if tt.err == nil {
					t.Fatalf("Unexpected error: '%v'", err)
				}
				t.Errorf("Expected error: '%v', got: '%v'", tt.err, err)
			}
			if err != nil {
				return
			}

			if s.Location != "s3://url" {
				t.Errorf("Expected s3://url, got %s", s.Location)
			}
		})
	}
}
