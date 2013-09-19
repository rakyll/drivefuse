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

package logger

import (
	"log"
	"os"
)

const (
	LogLevelFatal   = 0
	LogLevelVerbose = 1
	LogLevelDebug   = 2
)

var (
	logLevel = LogLevelFatal
)

func init() {
	levels := []string{"fatal", "verbose", "debug"}
	envLogLevel := os.Getenv("DRIVEFUSE_LOGLEVEL")
	if envLogLevel == "" {
		return
	}
	for index, val := range levels {
		if envLogLevel == val {
			logLevel = index
		}
	}
}

func F(args ...interface{}) {
	if logLevel >= LogLevelFatal {
		log.Fatalln(args...)
	}
}

func D(args ...interface{}) {
	if logLevel >= LogLevelDebug {
		log.Println(args...)
	}
}

func V(args ...interface{}) {
	if logLevel >= LogLevelVerbose {
		log.Println(args...)
	}
}
