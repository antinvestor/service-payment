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
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/antinvestor/service-payments/pkg/apperrors"
)

// QueryContainer represents the format of query subsection inside `must` or `should`.
type QueryContainer struct {
	Fields     []map[string]map[string]interface{} `json:"fields"`
	Terms      []map[string]interface{}            `json:"terms"`
	RangeItems []map[string]map[string]interface{} `json:"ranges"`
}

// SearchRawQuery represents the format of search query.
type SearchRawQuery struct {
	Offset int `json:"from,omitempty"`
	Limit  int `json:"size,omitempty"`
	Query  struct {
		MustClause   QueryContainer `json:"must"`
		ShouldClause QueryContainer `json:"should"`
	} `json:"query"`
}

// SearchSQLQuery holds information of a search SQL query.
type SearchSQLQuery struct {
	sql    string
	args   []interface{}
	offset int
	limit  int

	batchSize int
}

func (sq *SearchSQLQuery) canLoad() bool {
	return sq.offset < sq.limit
}

func (sq *SearchSQLQuery) stop(loadedCount int) bool {
	sq.offset += loadedCount
	if sq.offset+sq.batchSize > sq.limit {
		sq.batchSize = sq.limit - sq.offset
	}

	return loadedCount < sq.batchSize
}

func hasValidKeys(items interface{}) bool {
	var validKey = regexp.MustCompile(`^[a-z_A-Z.]+$`)
	switch t := items.(type) {
	case []map[string]interface{}:
		for _, item := range t {
			for key := range item {
				if !validKey.MatchString(key) {
					return false
				}
			}
		}
		return true
	case []map[string]map[string]interface{}:
		for _, item := range t {
			for key := range item {
				if !validKey.MatchString(key) {
					return false
				}
			}
		}
		return true
	default:
		return false
	}
}

// NewSearchRawQuery returns a new instance of SearchRawQuery.
func NewSearchRawQuery(_ context.Context, q string) (*SearchRawQuery, apperrors.ApplicationError) {
	rawQuery := new(SearchRawQuery)
	if err := json.Unmarshal([]byte(q), rawQuery); err != nil {
		return nil, apperrors.ErrSearchQueryHasInvalidFormat.Override(err)
	}

	if rawQuery.Query.MustClause.Fields == nil {
		rawQuery.Query.MustClause.Fields = []map[string]map[string]interface{}{}
	}
	if rawQuery.Query.MustClause.Terms == nil {
		rawQuery.Query.MustClause.Terms = []map[string]interface{}{}
	}
	if rawQuery.Query.MustClause.RangeItems == nil {
		rawQuery.Query.MustClause.RangeItems = []map[string]map[string]interface{}{}
	}
	if rawQuery.Query.ShouldClause.Fields == nil {
		rawQuery.Query.ShouldClause.Fields = []map[string]map[string]interface{}{}
	}
	if rawQuery.Query.ShouldClause.Terms == nil {
		rawQuery.Query.ShouldClause.Terms = []map[string]interface{}{}
	}
	if rawQuery.Query.ShouldClause.RangeItems == nil {
		rawQuery.Query.ShouldClause.RangeItems = []map[string]map[string]interface{}{}
	}

	checkList := []interface{}{
		rawQuery.Query.MustClause.Fields,
		rawQuery.Query.MustClause.Terms,
		rawQuery.Query.MustClause.RangeItems,
		rawQuery.Query.ShouldClause.Fields,
		rawQuery.Query.ShouldClause.Terms,
		rawQuery.Query.ShouldClause.RangeItems,
	}
	for _, item := range checkList {
		if !hasValidKeys(item) {
			return nil, apperrors.ErrSearchQueryHasInvalidKeys
		}
	}
	return rawQuery, nil
}

// ToQueryConditions converts a raw search query to SQL conditions.
func (rawQuery *SearchRawQuery) ToQueryConditions() *SearchSQLQuery {
	var conditionSQL string
	var conditionArgs []interface{}

	var mustWhere []string
	mustClause := rawQuery.Query.MustClause
	fieldsWhere, fieldsArgs := convertFieldsToSQL(mustClause.Fields)
	mustWhere = append(mustWhere, fieldsWhere...)
	conditionArgs = append(conditionArgs, fieldsArgs...)

	termsWhere, termsArgs := convertTermsToSQL(mustClause.Terms)
	mustWhere = append(mustWhere, termsWhere...)
	conditionArgs = append(conditionArgs, termsArgs...)

	rangesWhere, rangesArgs := convertRangesToSQL(mustClause.RangeItems)
	mustWhere = append(mustWhere, rangesWhere...)
	conditionArgs = append(conditionArgs, rangesArgs...)

	var shouldWhere []string
	shouldClause := rawQuery.Query.ShouldClause
	fieldsWhere, fieldsArgs = convertFieldsToSQL(shouldClause.Fields)
	shouldWhere = append(shouldWhere, fieldsWhere...)
	conditionArgs = append(conditionArgs, fieldsArgs...)

	termsWhere, termsArgs = convertTermsToSQL(shouldClause.Terms)
	shouldWhere = append(shouldWhere, termsWhere...)
	conditionArgs = append(conditionArgs, termsArgs...)

	rangesWhere, rangesArgs = convertRangesToSQL(shouldClause.RangeItems)
	shouldWhere = append(shouldWhere, rangesWhere...)
	conditionArgs = append(conditionArgs, rangesArgs...)

	if len(mustWhere) == 0 && len(shouldWhere) == 0 {
		return &SearchSQLQuery{
			sql:       conditionSQL,
			args:      conditionArgs,
			offset:    0,
			limit:     SystemBatchSize,
			batchSize: SystemBatchSize,
		}
	}

	if len(mustWhere) != 0 {
		conditionSQL += "(" + strings.Join(mustWhere, " AND ") + ")"
		if len(shouldWhere) != 0 {
			conditionSQL += " AND "
		}
	}

	if len(shouldWhere) != 0 {
		conditionSQL += "(" + strings.Join(shouldWhere, " OR ") + ")"
	}

	var offset = rawQuery.Offset
	var limit = rawQuery.Limit

	if offset <= 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	} else if limit > MaxSearchLimit {
		limit = MaxSearchLimit
	}

	batchSize := limit
	if SystemBatchSize < limit {
		batchSize = SystemBatchSize
	}

	return &SearchSQLQuery{sql: conditionSQL, args: conditionArgs, offset: offset, limit: limit, batchSize: batchSize}
}
