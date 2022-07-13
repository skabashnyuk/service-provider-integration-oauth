// Copyright (c) 2021 Red Hat, Inc.
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

//go:build !release
// +build !release

package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// fakeRoundTrip casts a function into a http.RoundTripper
type fakeRoundTrip func(r *http.Request) (*http.Response, error)

var _ http.RoundTripper = fakeRoundTrip(nil)

func (f fakeRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func GetEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvOrDefaultInt(key string, fallback int) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		result, err := strconv.Atoi(value)
		if err != nil {
			return fallback, fmt.Errorf("fail to parse value %w", err)
		}
		return result, nil
	}
	return fallback, nil
}
