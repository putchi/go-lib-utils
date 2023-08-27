package utils

import (
	"errors"
	"fmt"
	"github.com/elimity-com/scim"
	"github.com/labstack/gommon/log"
	"github.com/scim2/filter-parser/v2"
	"strconv"
	"strings"
)

type MappingValues struct {
	MappingValue string
	DataType     string
	IsSortable   bool
}

var scimOpMap = map[string]string{
	"eq": "=",
	"ne": "!=",
	"co": "LIKE",
	"sw": "LIKE",
	"ew": "LIKE",
	"pr": "IS NOT NULL",
	"gt": ">",
	"ge": ">=",
	"lt": "<",
	"le": "<=",
}

var matchingOpByScimOp = map[string][]string{
	"co": {"%", "%"},
	"sw": {"", "%"},
	"ew": {"%", ""},
}

type SqlQuery struct {
	fieldMappings map[filter.AttributePath]MappingValues
	// Filter should be a pointer
	//Filter     *string
	Filter     *strings.Builder
	Parameters map[int]interface{}
	Error      error
	requestId  string
	Limit      string
	OrderBy    string
}

func ParseScimParams(params scim.ListRequestParams, fieldMappings map[filter.AttributePath]MappingValues, orderField string, orderDirection string) (*SqlQuery, error) {

	offsetFactor := 1
	offset := 0
	if params.StartIndex > 0 {
		offsetFactor = params.StartIndex
		offset = offsetFactor * params.Count
	}

	sqlQuery := &SqlQuery{
		Parameters:    make(map[int]interface{}),
		Error:         nil,
		fieldMappings: fieldMappings,
		Limit:         fmt.Sprintf("limit %s, %s", strconv.Itoa(offset), strconv.Itoa(params.Count)),
		OrderBy:       fmt.Sprintf("order by %s %s", orderField, orderDirection),
	}

	if params.Filter == nil {
		return sqlQuery, nil
	}
	sqlQuery.Filter = &strings.Builder{}
	resp, err := sqlQuery.visitList(params.Filter)
	if err == nil {
		resp.fieldMappings = nil
	}
	return resp, err
}

func (sq *SqlQuery) GetParameterList() []interface{} {
	params := make([]interface{}, len(sq.Parameters))
	for k, v := range sq.Parameters {
		params[k] = v
	}
	return params
}

func (sq *SqlQuery) visitList(pFilter interface{}) (*SqlQuery, error) {

	switch v := pFilter.(type) {
	case *filter.LogicalExpression:
		sq.buildLogicalExpression("", pFilter.(*filter.LogicalExpression))
	case *filter.AttributeExpression:
		sq.buildAttributeExpression("", pFilter.(*filter.AttributeExpression), " ")
	case *filter.ValuePath:
		sq.buildValuePathExpression(pFilter.(*filter.ValuePath))
	default:
		sq.Error = errors.New(fmt.Sprintf("I don't know about type %T!", v))
	}
	if sq.Error != nil {
		return nil, sq.Error
	}
	return sq, nil
}

func (sq *SqlQuery) buildValuePathExpression(pFilter *filter.ValuePath) {
	switch pFilter.ValueFilter.(type) {
	case *filter.LogicalExpression:
		sq.buildLogicalExpression(pFilter.String(), pFilter.ValueFilter.(*filter.LogicalExpression))
	case *filter.AttributeExpression:
		sq.buildAttributeExpression(pFilter.String(), pFilter.ValueFilter.(*filter.AttributeExpression), " ")
	}
}

func (sq *SqlQuery) buildLogicalExpression(parent string, pFilter *filter.LogicalExpression) {
	// 1. check X
	_, _ = sq.Filter.WriteString(" ( ")
	switch pFilter.Left.(type) {
	case *filter.LogicalExpression:
		sq.buildLogicalExpression(parent, pFilter.Left.(*filter.LogicalExpression))
		_, _ = sq.Filter.WriteString(" " + string(pFilter.Operator) + " ")
	case *filter.AttributeExpression:
		sq.buildAttributeExpression(parent, pFilter.Left.(*filter.AttributeExpression), string(pFilter.Operator))
	}
	if sq.Error != nil {
		return
	}

	switch pFilter.Right.(type) {
	case *filter.LogicalExpression:
		sq.buildLogicalExpression(parent, pFilter.Right.(*filter.LogicalExpression))
	case *filter.AttributeExpression:
		sq.buildAttributeExpression(parent, pFilter.Right.(*filter.AttributeExpression), " ")
	}
	_, _ = sq.Filter.WriteString(" ) ")
}

func (sq *SqlQuery) buildAttributeExpression(parent string, pFilter *filter.AttributeExpression, token string) {
	sqlField, sqlFieldType, sqlOperator := "", "", ""
	var sqlValue interface{}
	// 1. find sql field
	for k, v := range sq.fieldMappings {
		if (parent == k.AttributeName && *k.SubAttribute == pFilter.AttributePath.AttributeName) ||
			(k.AttributeName == pFilter.AttributePath.AttributeName &&
				k.SubAttribute == pFilter.AttributePath.SubAttribute) {

			sqlField = v.MappingValue
			sqlFieldType = v.DataType
			break
		}
	}
	if sqlField == "" {
		log.Print(sq.requestId, "Invalid field supplied\n")
		sq.Error = errors.New("invalid/unmapped field supplied")
		return
	}
	sqlOperator = scimOpMap[strings.ToLower(string(pFilter.Operator))]

	valueWrapper, ok := matchingOpByScimOp[strings.ToLower(string(pFilter.Operator))]

	if string(pFilter.Operator) != "pr" {
		if !ok {
			// need to check type here
			switch sqlFieldType {
			case "int":
				sqlValue = pFilter.CompareValue.(int)
				//TODO: Consider this case to test
			case "int64":
				sqlValue = pFilter.CompareValue.(int64)
			case "bool":
				sqlValue = pFilter.CompareValue.(bool)
			default:
				sqlValue = pFilter.CompareValue
			}
		} else {
			sqlValue = valueWrapper[0] + pFilter.CompareValue.(string) + valueWrapper[1]
		}
		_, _ = sq.Filter.WriteString(fmt.Sprintf(" (%s %s ?) %s", sqlField, sqlOperator, strings.ToUpper(token)))
		sq.Parameters[len(sq.Parameters)] = sqlValue
	} else {
		_, _ = sq.Filter.WriteString(fmt.Sprintf(" (%s %s) %s", sqlField, sqlOperator, strings.ToUpper(token)))
	}

}
