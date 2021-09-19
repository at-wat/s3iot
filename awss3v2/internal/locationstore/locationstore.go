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

// Package locationstore provides s3.Client wrapper to store object location.
package locationstore

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LocationStore struct {
	s3.HTTPClient
	Location string
}

func (s *LocationStore) Do(req *http.Request) (*http.Response, error) {
	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return res, err
	}
	if res.Request != nil && res.Request.URL != nil {
		u := *res.Request.URL
		u.RawQuery = ""
		s.Location = u.String()
	}
	return res, err
}
