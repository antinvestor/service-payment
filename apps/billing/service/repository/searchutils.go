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

package repository

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	SystemBatchSize = 30
	MaxSearchLimit  = 10000
)

func jsonify(input interface{}) string {
	j, _ := json.Marshal(input)
	return string(j)
}

func sqlComparisonOp(op string) string {
	switch op {
	case "gt":
		return ">"
	case "lt":
		return "<"
	case "gte":
		return ">="
	case "lte":
		return "<="
	case "ne":
		return "!="
	case "like":
		return "LIKE"
	case "notlike":
		return "NOT LIKE"
	case "is":
		return "IS"
	case "isnot":
		return "IS NOT"
	case "in":
		return "IN"
	case "notin":
		return "NOT IN"
	}
	return "="
}

func convertTermsToSQL(terms []map[string]interface{}) ([]string, []interface{}) {
	where := []string{}
	args := []interface{}{}
	for _, term := range terms {
		var conditions []string
		for key, value := range term {
			conditions = append(
				conditions,
				fmt.Sprintf("data->'%s' @> ?::jsonb", key),
			)
			args = append(args, jsonify(value))
		}
		where = append(where, "("+strings.Join(conditions, " AND ")+")")
	}
	return where, args
}

func convertRangesToSQL(ranges []map[string]map[string]interface{}) ([]string, []interface{}) {
	where := []string{}
	args := []interface{}{}
	for _, rangeItem := range ranges {
		var conditions []string
		for key, comparison := range rangeItem {
			for op, value := range comparison {
				condn, arguments := getSQLConditionAndArgsFromRange(key, op, value)
				conditions = append(conditions, condn)
				for _, arg := range arguments {
					if arg != nil {
						args = append(args, arg)
					}
				}
			}
		}
		where = append(where, "("+strings.Join(conditions, " AND ")+")")
	}
	return where, args
}

func getSQLConditionAndArgsFromRange(key string, op string, value interface{}) (string, []interface{}) {
	getConditionAndArgs := func(key string, op string, val interface{}) (string, interface{}) {
		var condn string
		var arg interface{}
		switch val.(type) {
		case int, int8, int16, int32, int64, float32, float64:
			condn = fmt.Sprintf("(data->>'%s')::float %s ?", key, sqlComparisonOp(op))
			arg = val
		case nil:
			condn = fmt.Sprintf("data->>'%s' %s null", key, sqlComparisonOp(op))
			arg = nil
		default:
			condn = fmt.Sprintf("data->>'%s' %s ?", key, sqlComparisonOp(op))
			arg = val
		}
		return condn, arg
	}

	condition := ""
	args := []interface{}{}
	switch op {
	case "in", "nin":
		var opnew string
		if op == "in" {
			opnew = "eq"
		} else {
			opnew = "ne"
		}
		values, _ := value.([]interface{})
		for i, val := range values {
			c, arg := getConditionAndArgs(key, opnew, val)
			args = append(args, arg)
			if i == 0 {
				condition = c
			} else {
				condition = condition + " OR " + c
			}
		}
	default:
		c, arg := getConditionAndArgs(key, op, value)
		condition = c
		args = append(args, arg)
	}
	return condition, args
}

func convertFieldsToSQL(fields []map[string]map[string]interface{}) ([]string, []interface{}) {
	where := []string{}
	args := []interface{}{}
	for _, field := range fields {
		var conditions []string
		for key, comparison := range field {
			for op, value := range comparison {
				condn := fmt.Sprintf("%s %s ?", key, sqlComparisonOp(op))
				conditions = append(conditions, condn)
				args = append(args, value)
			}
		}
		where = append(where, "("+strings.Join(conditions, " AND ")+")")
	}
	return where, args
}
