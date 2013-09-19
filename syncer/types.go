// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncer

import (
	"metadata"
	driveclient "third_party/code.google.com/p/google-api-go-client/drive/v2"
)

type Syncer interface {
	// Initiates a new Syncer with an API service and
	// a metadata manager.
	New(*driveclient.Service, *metadata.MetaService) *Syncer

	// Starts a periodic syncing, returns immediately.
	Start()

	// Starts a sync if no syncing ,waits for the existing
	// sync process to be finished. Ignores incremental syncs
	// if isForce is set.
	Sync(isForce bool) (err error)
}
