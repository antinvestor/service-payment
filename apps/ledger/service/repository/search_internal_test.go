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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSearchEngine_ValidNamespaces(t *testing.T) {
	validNamespaces := []string{
		SearchNamespaceLedgers,
		SearchNamespaceAccounts,
		SearchNamespaceTransactions,
		SearchNamespaceTransactionEntries,
	}
	for _, ns := range validNamespaces {
		se, err := NewSearchEngine(nil, ns)
		require.NoError(t, err, "namespace %s should be valid", ns)
		assert.NotNil(t, se)
		assert.Equal(t, ns, se.namespace)
	}
}

func TestNewSearchEngine_InvalidNamespace(t *testing.T) {
	se, err := NewSearchEngine(nil, "invalid")
	require.Error(t, err)
	assert.Nil(t, se)
	assert.Contains(t, err.Error(), "not recognised")
}

func TestNewSearchRawQuery_InvalidJSON(t *testing.T) {
	_, err := NewSearchRawQuery(context.Background(), "not json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestNewSearchRawQuery_InvalidKeys(t *testing.T) {
	_, err := NewSearchRawQuery(
		context.Background(),
		`{"query": {"must": {"fields": [{"id; DROP TABLE": {"eq": "x"}}]}}}`,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid keys")
}

func TestNewSearchRawQuery_EmptyQuery(t *testing.T) {
	rq, err := NewSearchRawQuery(context.Background(), "{}")
	require.NoError(t, err)
	assert.NotNil(t, rq)
}

func TestNewSearchRawQuery_ValidQuery(t *testing.T) {
	rq, err := NewSearchRawQuery(context.Background(), `{"query": {"must": {"fields": [{"type": {"eq": "ASSET"}}]}}}`)
	require.NoError(t, err)
	assert.NotNil(t, rq)
	assert.Len(t, rq.Query.MustClause.Fields, 1)
}

func TestToQueryConditions_Empty(t *testing.T) {
	rq, err := NewSearchRawQuery(context.Background(), "{}")
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Empty(t, sq.sql)
	assert.Empty(t, sq.args)
	assert.Equal(t, SystemBatchSize, sq.batchSize)
}

func TestToQueryConditions_WithMustFields(t *testing.T) {
	rq, err := NewSearchRawQuery(context.Background(), `{"query": {"must": {"fields": [{"type": {"eq": "ASSET"}}]}}}`)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Contains(t, sq.sql, "type")
	assert.Len(t, sq.args, 1)
}

func TestToQueryConditions_WithShouldFields(t *testing.T) {
	rq, err := NewSearchRawQuery(
		context.Background(),
		`{"query": {"should": {"fields": [{"type": {"eq": "ASSET"}}, {"type": {"eq": "LIABILITY"}}]}}}`,
	)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Contains(t, sq.sql, "OR")
	assert.Len(t, sq.args, 2)
}

func TestToQueryConditions_WithTerms(t *testing.T) {
	rq, err := NewSearchRawQuery(context.Background(), `{"query": {"must": {"terms": [{"status": "completed"}]}}}`)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Contains(t, sq.sql, "data->")
	assert.Len(t, sq.args, 1)
}

func TestToQueryConditions_WithRanges(t *testing.T) {
	rq, err := NewSearchRawQuery(
		context.Background(),
		`{"query": {"must": {"ranges": [{"charge": {"gte": 2000, "lte": 4000}}]}}}`,
	)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.NotEmpty(t, sq.sql)
	assert.GreaterOrEqual(t, len(sq.args), 2)
}

func TestToQueryConditions_WithLimitAndOffset(t *testing.T) {
	rq, err := NewSearchRawQuery(
		context.Background(),
		`{"from": 10, "size": 50, "query": {"must": {"fields": [{"type": {"eq": "ASSET"}}]}}}`,
	)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Equal(t, 10, sq.offset)
	assert.Equal(t, 50, sq.limit)
}

func TestToQueryConditions_LimitCapping(t *testing.T) {
	rq, err := NewSearchRawQuery(
		context.Background(),
		`{"size": 999999, "query": {"must": {"fields": [{"type": {"eq": "ASSET"}}]}}}`,
	)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Equal(t, MaxSearchLimit, sq.limit)
}

func TestToQueryConditions_WithMustAndShould(t *testing.T) {
	rq, err := NewSearchRawQuery(context.Background(), `{
		"query": {
			"must": {"fields": [{"type": {"eq": "ASSET"}}]},
			"should": {"fields": [{"id": {"eq": "123"}}]}
		}
	}`)
	require.NoError(t, err)

	sq := rq.ToQueryConditions()
	assert.Contains(t, sq.sql, "AND")
}

func TestSearchSQLQuery_CanLoad(t *testing.T) {
	sq := &SearchSQLQuery{offset: 0, limit: 100}
	assert.True(t, sq.canLoad())

	sq.offset = 100
	assert.False(t, sq.canLoad())
}

func TestSearchSQLQuery_Stop(t *testing.T) {
	sq := &SearchSQLQuery{offset: 0, limit: 100, batchSize: 30}
	stopped := sq.stop(30)
	assert.False(t, stopped)
	assert.Equal(t, 30, sq.offset)

	stopped = sq.stop(10)
	assert.True(t, stopped)
}

func TestHasValidKeys_ValidFieldKeys(t *testing.T) {
	fields := []map[string]map[string]interface{}{
		{"type": {"eq": "ASSET"}},
		{"ledger_id": {"eq": "123"}},
		{"data.name": {"eq": "test"}},
	}
	assert.True(t, hasValidKeys(fields))
}

func TestHasValidKeys_InvalidFieldKeys(t *testing.T) {
	fields := []map[string]map[string]interface{}{
		{"type; DROP TABLE": {"eq": "ASSET"}},
	}
	assert.False(t, hasValidKeys(fields))
}

func TestHasValidKeys_ValidTermKeys(t *testing.T) {
	terms := []map[string]interface{}{
		{"status": "completed"},
	}
	assert.True(t, hasValidKeys(terms))
}

func TestHasValidKeys_InvalidTermKeys(t *testing.T) {
	terms := []map[string]interface{}{
		{"status' OR 1=1": "completed"},
	}
	assert.False(t, hasValidKeys(terms))
}

func TestHasValidKeys_UnsupportedType(t *testing.T) {
	assert.False(t, hasValidKeys("unsupported"))
}

func TestSqlComparisonOp(t *testing.T) {
	tests := []struct {
		op       string
		expected string
	}{
		{"gt", ">"},
		{"lt", "<"},
		{"gte", ">="},
		{"lte", "<="},
		{"ne", "!="},
		{"like", "LIKE"},
		{"notlike", "NOT LIKE"},
		{"is", "IS"},
		{"isnot", "IS NOT"},
		{"in", "IN"},
		{"notin", "NOT IN"},
		{"eq", "="},
		{"unknown", "="},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, sqlComparisonOp(tc.op), "op=%s", tc.op)
	}
}

func TestGetSQLConditionAndArgsFromRange_Numeric(t *testing.T) {
	condition, args := getSQLConditionAndArgsFromRange("charge", "gte", float64(2000))
	assert.Contains(t, condition, "float")
	assert.Contains(t, condition, ">=")
	assert.Len(t, args, 1)
}

func TestGetSQLConditionAndArgsFromRange_String(t *testing.T) {
	condition, args := getSQLConditionAndArgsFromRange("date", "gt", "2017-01-01")
	assert.Contains(t, condition, ">")
	assert.NotContains(t, condition, "float")
	assert.Len(t, args, 1)
}

func TestGetSQLConditionAndArgsFromRange_Nil(t *testing.T) {
	condition, args := getSQLConditionAndArgsFromRange("field", "is", nil)
	assert.Contains(t, condition, "IS")
	assert.Contains(t, condition, "null")
	assert.Len(t, args, 1)
	assert.Nil(t, args[0])
}

func TestGetSQLConditionAndArgsFromRange_In(t *testing.T) {
	values := []interface{}{"a", "b", "c"}
	condition, args := getSQLConditionAndArgsFromRange("status", "in", values)
	assert.Contains(t, condition, "OR")
	assert.Len(t, args, 3)
}

func TestGetSQLConditionAndArgsFromRange_Nin(t *testing.T) {
	values := []interface{}{"x", "y"}
	condition, args := getSQLConditionAndArgsFromRange("status", "nin", values)
	assert.Contains(t, condition, "!=")
	assert.Contains(t, condition, "OR")
	assert.Len(t, args, 2)
}

func TestConvertFieldsToSQL(t *testing.T) {
	fields := []map[string]map[string]interface{}{
		{"type": {"eq": "ASSET"}},
	}
	where, args := convertFieldsToSQL(fields)
	assert.Len(t, where, 1)
	assert.Contains(t, where[0], "type = ?")
	assert.Len(t, args, 1)
}

func TestConvertTermsToSQL(t *testing.T) {
	terms := []map[string]interface{}{
		{"status": "completed"},
	}
	where, args := convertTermsToSQL(terms)
	assert.Len(t, where, 1)
	assert.Contains(t, where[0], "data->'status'")
	assert.Len(t, args, 1)
}

func TestConvertRangesToSQL(t *testing.T) {
	ranges := []map[string]map[string]interface{}{
		{"charge": {"gte": float64(2000), "lte": float64(4000)}},
	}
	where, args := convertRangesToSQL(ranges)
	assert.Len(t, where, 1)
	assert.GreaterOrEqual(t, len(args), 2)
}

func TestJoinAND(t *testing.T) {
	result := joinAND([]string{"a = 1", "b = 2", "c = 3"})
	assert.Equal(t, "a = 1 AND b = 2 AND c = 3", result)
}

func TestJoinAND_Single(t *testing.T) {
	result := joinAND([]string{"x = 1"})
	assert.Equal(t, "x = 1", result)
}
