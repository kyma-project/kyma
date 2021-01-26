/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package normalizator

import (
	"regexp"
	"strings"
)

// defaultNormalizationPrefix is a fixed string used as prefix during default normalization
const defaultNormalizationPrefix = "mp-"

type DefaultNormalizator struct{}

// DefaultNormalizer is the default normalization function used as normalization function;
// It attempts to validate if the input is already normalized so it doesn't apply normalization again.
func Normalize(name string) string {
	if isNormalized(name) {
		return name
	}

	prefixedName := defaultNormalizationPrefix + name
	return normalize(prefixedName)
}

func normalize(name string) string {
	normalizedName := strings.ToLower(name)
	normalizedName = regexp.MustCompile("[^-a-z0-9]").ReplaceAllString(normalizedName, "-")
	normalizedName = regexp.MustCompile("-{2,}").ReplaceAllString(normalizedName, "-")
	normalizedName = regexp.MustCompile("-$").ReplaceAllString(normalizedName, "")

	return normalizedName
}

func isNormalized(name string) bool {
	if !strings.HasPrefix(name, defaultNormalizationPrefix) {
		return false
	}
	normalizedName := normalize(name)
	return name == normalizedName
}
