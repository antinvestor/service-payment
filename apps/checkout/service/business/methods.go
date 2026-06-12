// Copyright 2023-2026 Ant Investor Ltd
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

package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Method describes one renderable payment method and its prompt route hint.
type Method struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Route      string   `json:"route"`
	Prefixes   []string `json:"prefixes"`
	Currencies []string `json:"currencies"`
}

// MethodRegistry is the config-defined list of supported methods.
type MethodRegistry struct {
	Methods []Method
}

// ParseMethodRegistry parses the CHECKOUT_METHODS config JSON.
func ParseMethodRegistry(raw string) (*MethodRegistry, error) {
	var methods []Method
	if err := json.Unmarshal([]byte(raw), &methods); err != nil {
		return nil, fmt.Errorf("parse method registry: %w", err)
	}
	if len(methods) == 0 {
		return nil, errors.New("method registry is empty")
	}
	return &MethodRegistry{Methods: methods}, nil
}

// Available returns methods filtered by an optional restriction key list.
func (r *MethodRegistry) Available(restriction []string) []Method {
	if len(restriction) == 0 {
		return r.Methods
	}
	allowed := make(map[string]bool, len(restriction))
	for _, k := range restriction {
		allowed[k] = true
	}
	var out []Method
	for _, m := range r.Methods {
		if allowed[m.Key] {
			out = append(out, m)
		}
	}
	return out
}

// Get returns the method for a key, or false.
func (r *MethodRegistry) Get(key string) (Method, bool) {
	for _, m := range r.Methods {
		if m.Key == key {
			return m, true
		}
	}
	return Method{}, false
}

// Preselect picks the default method: profile clue -> phone prefix -> first.
// methods must be non-empty.
func Preselect(methods []Method, clueKey, phoneNumber string) Method {
	for _, m := range methods {
		if clueKey != "" && m.Key == clueKey {
			return m
		}
	}
	phone := strings.TrimPrefix(strings.TrimSpace(phoneNumber), "+")
	if phone != "" {
		for _, m := range methods {
			for _, p := range m.Prefixes {
				if strings.HasPrefix(phone, p) {
					return m
				}
			}
		}
	}
	return methods[0]
}
