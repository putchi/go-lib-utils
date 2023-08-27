package utils

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMappingBuilder_retrieveValueFromDoc(t *testing.T) {

	jsonString := `{
	"foo": 1,
	"bar": 2,
	"test": "Hello, world!",
	"baz": 123.1,
	"array": [
		{"foo": 1},
		{"bar": 2},
		{"baz": 3}
	],
	"subobj": {
		"sfoo": "1",
		"foo": 1,
		"subarray": [1,2,3],
		"subsubobj": {
			"bar": 2,
			"baz": 3,
			"array": ["hello", "world"]
		}
	},
	"bool": true
	}`
	doc := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(jsonString))
	dec.Decode(&doc)

	type args struct {
		mappings map[string]string
		doc      map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Good",
			args: args{
				mappings: map[string]string{
					"field1": "foo",
					"field2": "subobj.sfoo",
				},
				doc: doc,
			},
			want: map[string]interface{}{
				"field1": "1",
				"field2": "1",
			},
		},
		{
			name: "Simple join",
			args: args{
				mappings: map[string]string{
					"field1": "foo+subobj.sfoo",
				},
				doc: doc,
			},
			want: map[string]interface{}{
				"field1": "11",
			},
		},
		{
			name: "Simple join with constant",
			args: args{
				mappings: map[string]string{
					"field1": "foo+'subobj.sfoo",
				},
				doc: doc,
			},
			want: map[string]interface{}{
				"field1": "1subobj.sfoo",
			},
		},
		{
			name: "Join with constant in center",
			args: args{
				mappings: map[string]string{
					"field1": "foo+' +subobj.sfoo",
				},
				doc: doc,
			},
			want: map[string]interface{}{
				"field1": "1 1",
			},
		},
		{
			name: "Invalid field",
			args: args{
				mappings: map[string]string{
					"field1": "fooqwaszx",
				},
				doc: doc,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mb := NewMappingBuilder(tt.args.mappings, tt.args.doc)
			got, err := mb.BuildBody()
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCreateRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMappingBuilder_getPrefix(t *testing.T) {

	type args struct {
		source string
		sep    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Prefix 1",
			args: args{"prefix1:data", ":"},
			want: "prefix1",
		},
		{
			name: "Prefix empty",
			args: args{"data", ":"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &MappingBuilder{}
			if got := mb.getPrefix(tt.args.source, tt.args.sep); got != tt.want {
				t.Errorf("getPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMappingBuilder_buildBody(t *testing.T) {

	docData := "{\"city\":\"Nairobi\",\"contact\":{\"email\":\"no.where.125@lost.found125.com\",\"first_name\":\"Carmen\",\"last_name\":\"Sandiego\",\"phone\":\"(2561 123 123)\",\"title\":\"Madam\"},\"country\":\"Kenya\",\"merchant_name\":\"Eugene Test 125\",\"state_data\":{\"Crm_Account_Creation\":{\"Accounts\":{\"code\":\"SUCCESS\",\"details\":{\"Created_By\":{\"id\":\"48040000137642006\",\"name\":\"Eugene Petersen\"},\"Created_Time\":\"2021-11-24T12:47:48+02:00\",\"Modified_By\":{\"id\":\"48040000137642006\",\"name\":\"Eugene Petersen\"},\"Modified_Time\":\"2021-11-24T12:47:48+02:00\",\"id\":\"48040000145422050\"},\"message\":\"record added\",\"status\":\"success\"},\"Contacts\":{\"code\":\"SUCCESS\",\"details\":{\"Created_By\":{\"id\":\"48040000137642006\",\"name\":\"Eugene Petersen\"},\"Created_Time\":\"2021-11-24T12:47:49+02:00\",\"Modified_By\":{\"id\":\"48040000137642006\",\"name\":\"Eugene Petersen\"},\"Modified_Time\":\"2021-11-24T12:47:49+02:00\",\"id\":\"48040000145552001\"},\"message\":\"record added\",\"status\":\"success\"}},\"Map_Account_Creation\":{\"MerchantOnBoarding\":{\"CompanyToken\":\"32C57936-FD9E-4A7E-B3BC-1D564401AB70\",\"Result\":\"000\",\"ResultExplanation\":\"Merchant has been created\"}},\"Map_User_Creation\":{\"CreateUser\":{\"CompanyToken\":\"\",\"Result\":\"999\",\"ResultExplanation\":\"No permissions - contact support team.\"}}},\"type_of_business\":\"E-Commerce\",\"website\":\"https://you.found.me\"}"
	doc := make(map[string]interface{})
	err := json.Unmarshal([]byte(docData), &doc)
	if err != nil {
		t.Errorf("Unmarshal doc error:%s, doc:%s ", err.Error(), docData)
		return
	}

	type fields struct {
		mappings map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{

		{
			name: "test strategy default",
			fields: fields{
				mappings: map[string]string{"field1": "country"},
			},
			want: map[string]interface{}{"field1": "Kenya"},
		},
		{
			name: "test strategy default wrong key",
			fields: fields{
				mappings: map[string]string{"field1": "wrongkey"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test strategy iso3166_1",
			fields: fields{
				mappings: map[string]string{"iso3166_1:field1": "country"},
			},
			want: map[string]interface{}{"field1": "KE"},
		},
		{
			name: "test strategy iso3166_1 wrong key",
			fields: fields{
				mappings: map[string]string{"iso3166_1:field1": "wrongkey"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test strategy const",
			fields: fields{
				mappings: map[string]string{"const:field1": "const string"},
			},
			want: map[string]interface{}{"field1": "const string"},
		},
		{
			name: "test strategy templateVar 1",
			fields: fields{
				mappings: map[string]string{"templateVar:field1": "country"},
			},
			want: map[string]interface{}{"templateVars": map[string]string{
				"field1": "Kenya",
			}},
		},
		{
			name: "test strategy templateVar 2",
			fields: fields{
				mappings: map[string]string{
					"templateVar:field1": "country",
					"templateVar:field2": "city",
				},
			},
			want: map[string]interface{}{"templateVars": map[string]string{
				"field1": "Kenya",
				"field2": "Nairobi",
			}},
		},
		{
			name: "test strategy templateVar wrongkey",
			fields: fields{
				mappings: map[string]string{"templateVar:field1": "wrongkey"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test strategy only numbers",
			fields: fields{
				mappings: map[string]string{"onlynumbers:field1": "contact.phone"},
			},
			want: map[string]interface{}{"field1": "2561123123"},
		},
		{
			name: "test strategy only numbers wrong key",
			fields: fields{
				mappings: map[string]string{"onlynumbers:field1": "wrongkey"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test strategy array",
			fields: fields{
				mappings: map[string]string{"arr:field1": "city,country"},
			},
			want: map[string]interface{}{"field1": []string{"Nairobi", "Kenya"}},
		},
		{
			name: "test strategy array wrong key",
			fields: fields{
				mappings: map[string]string{"arr:field1": "wrongkey"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := NewMappingBuilder(tt.fields.mappings, doc)
			got, err := mb.BuildBody()
			if (err != nil) != tt.wantErr {
				t.Errorf("buildBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildBody() got = %v, want %v", got, tt.want)
			}
		})
	}
}
