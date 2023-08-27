package utils

import (
	"github.com/elimity-com/scim"
	"github.com/scim2/filter-parser/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessor_GetSqlQuery(t *testing.T) {
	var TransactionAttemptsMappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}: {"tra.id", "int", true},
	}
	expression, err := filter.ParseFilter([]byte("id pr"))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, TransactionAttemptsMappings, "tra.id", "asc")
	assert.NotNil(t, got)
	assert.NoError(t, err)
}

func TestProcessor_GetSqlQuery_UnknownField(t *testing.T) {
	var Mappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}: {"tra.id", "int", true},
	}
	expression, err := filter.ParseFilter([]byte("count pr"))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, Mappings, "tra.id", "asc")
	assert.Nil(t, got)
	assert.Error(t, err)
}

func TestProcessor_GetSqlQuery_Logical(t *testing.T) {
	var Mappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}:   {"tra.id", "int", true},
		filter.AttributePath{AttributeName: "cost"}: {"tra.cost", "int", true},
	}
	expression, err := filter.ParseFilter([]byte("id pr and cost pr"))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, Mappings, "tra.id", "asc")
	assert.NotNil(t, got)
	assert.NoError(t, err)
}

func TestProcessor_GetSqlQuery_LogicalError(t *testing.T) {
	var Mappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}:   {"tra.id", "int", true},
		filter.AttributePath{AttributeName: "cost"}: {"tra.cost", "int", true},
	}
	expression, err := filter.ParseFilter([]byte("id pr ans cost pr"))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, Mappings, "tra.id", "asc")
	assert.NotNil(t, got)
	assert.NoError(t, err)
}

func TestProcessor_GetSqlQuery_Value(t *testing.T) {
	var Mappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}:   {"tra.id", "int", true},
		filter.AttributePath{AttributeName: "cost"}: {"tra.cost", "int", true},
	}
	expression, err := filter.ParseFilter([]byte(`id pr and cost eq 300`))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, Mappings, "tra.id", "asc")
	assert.NotNil(t, got)
	assert.NoError(t, err)
}

func TestProcessor_GetSqlQuery_PathExpr(t *testing.T) {
	var Mappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}:     {"tra.id", "int", true},
		filter.AttributePath{AttributeName: "cost"}:   {"tra.cost", "int", true},
		filter.AttributePath{AttributeName: "emails"}: {"tra.emails", "string", true},
		filter.AttributePath{AttributeName: "enable"}: {"tra.enable", "bool", true},
	}
	expression, err := filter.ParseFilter([]byte(`id eq 1 and (cost ge 4 or emails co "example.org" or enable eq true)`))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, Mappings, "tra.id", "asc")
	assert.NotNil(t, got)
	assert.NoError(t, err)
}

func TestProcessor_GetSqlQuery_URN(t *testing.T) {
	var Mappings = map[filter.AttributePath]MappingValues{
		filter.AttributePath{AttributeName: "id"}:     {"tra.id", "int", true},
		filter.AttributePath{AttributeName: "cost"}:   {"tra.cost", "int", true},
		filter.AttributePath{AttributeName: "emails"}: {"tra.emails", "string", true},
		filter.AttributePath{AttributeName: "enable"}: {"tra.enable", "bool", true},
	}
	expression, err := filter.ParseFilter([]byte(`urn:ietf:params:scim:schemas:core:2.0:User:emails sw "a"`))
	listRequestParams := scim.ListRequestParams{
		Filter:     expression,
		Count:      10,
		StartIndex: 1,
	}
	got, err := ParseScimParams(listRequestParams, Mappings, "tra.id", "asc")
	assert.NotNil(t, got)
	assert.NoError(t, err)
}

func TestProcessor_GetParameterList(t *testing.T) {
	sqlQuery := &SqlQuery{
		Parameters: make(map[int]interface{}),
	}
	sqlQuery.Parameters[0] = "0"
	sqlQuery.Parameters[1] = "1"
	sqlQuery.Parameters[2] = "2"
	got := sqlQuery.GetParameterList()

	assert.Equal(t, []interface{}{"0", "1", "2"}, got)
}
