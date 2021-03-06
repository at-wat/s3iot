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

package s3iot

import (
	"reflect"
	"sort"
	"testing"

	"github.com/at-wat/s3iot/s3api"
)

func TestCompletedParts(t *testing.T) {
	part := func(e string, i int64) *s3api.CompletedPart {
		return &s3api.CompletedPart{ETag: &e, PartNumber: &i}
	}
	ps := completedParts{
		part("a", 1),
		part("c", 3),
		part("b", 2),
	}
	expected := completedParts{
		part("a", 1),
		part("b", 2),
		part("c", 3),
	}
	sort.Sort(ps)

	if !reflect.DeepEqual(expected, ps) {
		t.Errorf("Expected: %v, got %v", expected, ps)
	}
}
