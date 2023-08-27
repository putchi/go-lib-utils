package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type MappingStrategy interface {
	extractBodyValue(mapKey string, mapValue string) (value interface{}, err error)
}

type MappingBuilder struct {
	docJQ           *JsonQuery
	mappingStrategy MappingStrategy
	mappings        map[string]string
	body            map[string]interface{}
}

func NewMappingBuilder(mappings map[string]string, doc map[string]interface{}) *MappingBuilder {
	return &MappingBuilder{
		docJQ:    NewJsonQuery(doc),
		mappings: mappings,
		body:     make(map[string]interface{}),
	}
}

func (mb *MappingBuilder) BuildBody() (map[string]interface{}, error) {

	for mapKey, mapValue := range mb.mappings {
		bodyKey, bodyValue, err := mb.extractBodyData(mapKey, mapValue)
		if err != nil {
			return nil, err
		}

		mb.body[*bodyKey] = bodyValue
	}

	return mb.body, nil
}

func (mb *MappingBuilder) extractBodyData(mapKey string, mapValue string) (bodyKey *string, bodyValue interface{}, err error) {

	prefix := mb.getPrefix(mapKey, ":")
	bodyKey = StringPtr(strings.TrimPrefix(mapKey, prefix+":"))

	switch prefix {
	case "arr":
		mb.setMappingStrategy(&arrStrategy{mb})
	case "templateVar":
		bodyKey = StringPtr("templateVars")
		mb.setMappingStrategy(&templateVarStrategy{mb})
	case "const":
		mb.setMappingStrategy(&constStrategy{mb})
	case "iso3166_1":
		mb.setMappingStrategy(&iso3166Strategy{mb})
	case "onlynumbers":
		mb.setMappingStrategy(&onlynumbersStrategy{mb})
	case "idMerchantType":
		mb.setMappingStrategy(&idMerchantTypeStrategy{mb})
	default:
		mb.setMappingStrategy(&defaultStrategy{mb})
	}

	bodyValue, err = mb.mappingStrategy.extractBodyValue(mapKey, mapValue)
	if err != nil {
		return nil, nil, err
	}

	return bodyKey, bodyValue, nil
}

func (mb *MappingBuilder) setMappingStrategy(ms MappingStrategy) {
	mb.mappingStrategy = ms
}

func (mb *MappingBuilder) getPrefix(source string, sep string) string {
	slice := strings.Split(source, sep)
	if len(slice) > 1 {
		return slice[0]
	}
	return ""
}

func (mb *MappingBuilder) retrieveValueFromDoc(keys string) (string, error) {
	// is the keys a simple expression
	// constant values are prefixed with single-quote "'"
	splits := strings.Split(keys, "+")
	result := strings.Builder{}
	for _, key := range splits {
		if strings.HasPrefix(key, "'") {
			_, _ = result.WriteString(TrimLeftChars(key, 1))
		} else {
			keySplit := strings.Split(key, ".")
			i, err := mb.docJQ.Interface(keySplit...)
			if err != nil {
				return "", err
			}
			result.WriteString(fmt.Sprintf("%v", i))
		}
	}
	return result.String(), nil
}

// ////////////////////////////////////////////////////
// MAPPING STRATEGIES
// ////////////////////////////////////////////////////
type defaultStrategy struct {
	mb *MappingBuilder
}

func (s *defaultStrategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {

	value, err = s.mb.retrieveValueFromDoc(mapValue)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// ////////////////////////////////////////////////////
type arrStrategy struct {
	mb *MappingBuilder
}

func (s *arrStrategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {
	var arr []string
	elems := strings.Split(mapValue, ",")
	for _, elem := range elems {
		found, err := s.mb.retrieveValueFromDoc(elem)
		if err != nil {
			return nil, fmt.Errorf("error retrieving array type value: %v", err)
		}
		arr = append(arr, found)
	}
	return arr, nil
}

// ////////////////////////////////////////////////////
type templateVarStrategy struct {
	mb *MappingBuilder
}

func (s *templateVarStrategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {

	keyT := strings.TrimPrefix(mapKey, "templateVar:")
	valT, err := s.mb.retrieveValueFromDoc(mapValue)
	if err != nil {
		return nil, err
	}

	_, ok := s.mb.body["templateVars"]
	if !ok {
		s.mb.body["templateVars"] = map[string]string{}
	}

	templateVars := s.mb.body["templateVars"].(map[string]string)
	templateVars[keyT] = valT
	s.mb.body["templateVars"] = templateVars

	return s.mb.body["templateVars"], nil
}

// ////////////////////////////////////////////////////
type constStrategy struct {
	mb *MappingBuilder
}

func (s *constStrategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {
	return mapValue, nil
}

// ////////////////////////////////////////////////////
type iso3166Strategy struct {
	mb *MappingBuilder
}

func (s *iso3166Strategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {

	country, err := s.mb.retrieveValueFromDoc(mapValue)
	if err != nil {
		return nil, err
	}

	return GetISO3361CodeCountry(country), nil
}

// ////////////////////////////////////////////////////
type onlynumbersStrategy struct {
	mb *MappingBuilder
}

func (s *onlynumbersStrategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {

	re := regexp.MustCompile("\\D")
	bodyValue, err := s.mb.retrieveValueFromDoc(mapValue)
	if err != nil {
		return nil, fmt.Errorf("onlynumbersStrategy error: %v", err)
	}

	return re.ReplaceAllString(bodyValue, ""), nil
}

// ////////////////////////////////////////////////////
type idMerchantTypeStrategy struct {
	mb *MappingBuilder
}

func (s *idMerchantTypeStrategy) extractBodyValue(mapKey string, mapValue string) (value interface{}, err error) {

	name, err := s.mb.retrieveValueFromDoc(mapValue)
	if err != nil {
		return nil, err
	}

	return s.getId(name), nil
}

func (s *idMerchantTypeStrategy) getId(name string) int {
	return map[string]int{
		"Ecommerce":            1,
		"Hospitality":          2,
		"Insurance":            3,
		"Schools":              4,
		"Travel and Tourism":   5,
		"ICT":                  6,
		"SMEs":                 7,
		"Conferences & events": 8,
		"Activity Providers":   9,
		"Hospitals":            10,
	}[name]
}
