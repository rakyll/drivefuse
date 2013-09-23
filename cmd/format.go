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

// Contains formatting for the command-line user interface.
package cmd

// Bold generates ansi-escaped bold text.
func Bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

// Blue generates ansi-escaped blue text.
func Blue(str string) string {
	return Bold("\033[34m" + str + "\033[0m")
}
