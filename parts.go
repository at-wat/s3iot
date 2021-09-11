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

// Package s3iot provides S3 uploader applicable for unreliable and congestible network.
// Object can be uploaded with retry, pause/resume, and bandwidth limit.
package s3iot

type completedParts []*CompletedPart

func (a completedParts) Len() int {
	return len(a)
}

func (a completedParts) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a completedParts) Less(i, j int) bool {
	return *a[i].PartNumber < *a[j].PartNumber
}
