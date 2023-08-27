package utils

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGetRequestBody(t *testing.T) {
	buf := []byte(`{"msg":"Hello there!"}`)
	request, _ := http.NewRequest(http.MethodPost,
		"https://example.com/api/manage/secretreload",
		bytes.NewBuffer(buf))
	data, err := GetRequestBody(request)
	if err != nil {
		t.Errorf("TestGetRequestBody: failed with error %v", err)
	}

	if len(buf) != len(data) {
		t.Errorf("TestGetRequestBody: failed expected length %v, but got %v", len(buf), len(data))
	}

}

func TestGetRequestBodyError(t *testing.T) {
	request, _ := http.NewRequest(http.MethodPost,
		"https://example.com/api/manage/secretreload",
		nil)
	_, err := GetRequestBody(request)
	if err == nil {
		t.Errorf("TestGetRequestBodyError: expected failure, but got none")
		return
	}
	if err.Error() != "no body found in request" {
		t.Errorf("TestGetRequestBodyError: expected failure with message: \"%v\", but got \"%v\"",
			"no body found in request", err.Error())
	}

}

func TestFileExists(t *testing.T) {
	result := FileExists(os.Args[0])
	if !result {
		t.Errorf("TestFileExists success: expected %v, but got %v", true, result)
	}

	result = FileExists(os.Args[0] + "1")
	if result {
		t.Errorf("TestFileExists not found: expected %v, but got %v", false, result)
	}
}

func TestErrorToText(t *testing.T) {
	resp := &GenericResponse{
		Result:            "result",
		ResultExplanation: "result_explanation",
	}
	expectedResult := "result;result_explanation;;;"
	result := ErrorToText(*resp)
	if result != expectedResult {
		t.Errorf("Invalid result: expected[%s], got[%s]",
			expectedResult, result)
	}
}

func TestResponseWriterTextHtml(t *testing.T) {
	response := "{\"Message\":\"Hello World!\"}"
	responseWriter := httptest.NewRecorder()
	ResponseWriterTextHtml(responseWriter, response)
	if responseWriter.Header().Get("Content-Type") != ContentTypeApplicationXML {
		t.Errorf("Invalid Content-Type expected[%s], got[%s]",
			ContentTypeApplicationXML, responseWriter.Header().Get("Content-Type"))
	}
	if response != responseWriter.Body.String() {
		t.Errorf("Invalid body data: expecting[%s], got[%s]", response,
			responseWriter.Body.String())
	}
}

func TestResponseWriterXML(t *testing.T) {
	response := &struct {
		XMLName *xml.Name `xml:"root" json:",omitempty"`
		Message string    `json:"message"`
	}{
		XMLName: &xml.Name{},
		Message: "Hello World!",
	}
	expected := "<?xml version=\"1.0\" encoding=\"utf-8\"?><root><Message>Hello World!</Message></root>"
	responseWriter := httptest.NewRecorder()
	ResponseWriterXML(responseWriter, response)
	if responseWriter.Header().Get("Content-Type") != ContentTypeApplicationXML {
		t.Errorf("Invalid Content-Type expected[%s], got[%s]",
			ContentTypeApplicationXML, responseWriter.Header().Get("Content-Type"))
	}
	if expected != responseWriter.Body.String() {
		t.Errorf("Invalid response returning, expected[%s], got[%s]", expected,
			responseWriter.Body.String())
	}
}

func TestResponseWriterJSON(t *testing.T) {
	response := &struct {
		Message string `json:"message"`
	}{
		Message: "Hello World!",
	}
	expected := "{\"message\":\"Hello World!\"}\n"

	responseWriter := httptest.NewRecorder()
	ResponseWriterJSON(responseWriter, response)
	if responseWriter.Header().Get("Content-Type") != ContentTypeApplicationJSON {
		t.Errorf("Invalid Content-Type expected[%s], got[%s]",
			ContentTypeApplicationJSON, responseWriter.Header().Get("Content-Type"))
	}
	if expected != responseWriter.Body.String() {
		t.Errorf("Invalid body data: expecting[%s], got[%s]", expected,
			responseWriter.Body.String())
	}
}

func TestContains(t *testing.T) {
	search := []string{"findme", "mehere", "herefind"}
	findWhat := "mehere"
	result := Contains(search, findWhat)
	if !result {
		t.Errorf("Failed to find[%s]: expected[%v], got[%v]", findWhat, true, result)
	}
	findWhat = "here"
	result = Contains(search, findWhat)
	if result {
		t.Errorf("Failed to find[%s]: expected[%v], got[%v]", findWhat, true, result)
	}
}

func TestIsTimeType(t *testing.T) {
	valueType := reflect.TypeOf("str")
	if IsTimeType(valueType) {
		t.Errorf("Expected type inspection to be false")
	}
	valueType = reflect.TypeOf(time.Now())
	if !IsTimeType(valueType) {
		t.Errorf("Expected type inspection to be true")
	}
}

func TestIsCustomType(t *testing.T) {
	valueType := reflect.TypeOf(os.File{})
	result := IsCustomType(valueType)
	if result {
		t.Error("Invalid response, expecting[false], got[true]")
	}

	valueType = reflect.TypeOf(reflect.Method{})
	result = IsCustomType(valueType)
	if !result {
		t.Error("Invalid response, expecting[true], got[false]")
	}

	type A struct {
		Number string
	}
	valueType = reflect.TypeOf(A{})
	result = IsCustomType(valueType)
	if !result {
		t.Error("Invalid response, expecting[true], got[false]")
	}

	valueType = reflect.TypeOf(struct{ i int }{})
	result = IsCustomType(valueType)
	if result {
		t.Error("Invalid response, expecting[true], got[false]")
	}

	valueType = reflect.TypeOf(struct{ Abc A }{})
	result = IsCustomType(valueType)
	if !result {
		t.Error("Invalid response, expecting[true], got[false]")
	}

	valueType = reflect.TypeOf(errors.New(""))
	result = IsCustomType(valueType)
	if !result {
		t.Error("Invalid response, expecting[true], got[false]")
	}

	valueType = reflect.TypeOf([]byte{})
	result = IsCustomType(valueType)
	if result {
		t.Error("Invalid response, expecting[false], got[true]")
	}

}

func TestIndirect(t *testing.T) {
	type A struct {
		Number string
	}
	result := Indirect(reflect.ValueOf(&A{}))
	if result.Kind() != reflect.Struct {
		t.Errorf("Invalid Kind got returned, expected[%s], got[%s]", reflect.Struct.String(),
			result.Kind().String())
	}

	result = Indirect(Ptr(reflect.ValueOf(&A{})))
	if result.Kind() != reflect.Struct {
		t.Errorf("Invalid Kind got returned, expected[%s], got[%s]", reflect.Struct.String(),
			result.Kind().String())
	}
}

func TestIndirectType(t *testing.T) {
	type A struct {
		Number string
	}
	result := IndirectType(reflect.ValueOf(&A{}).Type())

	if result.Kind() != reflect.Struct {
		t.Errorf("Invalid Kind got returned, expected[%s], got[%s]", reflect.Struct.String(),
			result.Kind().String())
	}

}

func TestIsNil(t *testing.T) {
	type A struct {
		Number string
	}

	if !IsNil(nil) {
		t.Error("IsNil test failed, expect[true], got[false]")
	}

	if IsNil(10) {
		t.Error("IsNil test failed, expect[false], got[true]")
	}

	value := &A{}
	if IsNil(value) {
		t.Error("IsNil test failed, expect[false], got[true]")
	}

	if IsNil(*value) {
		t.Error("IsNil test failed, expect[false], got[true]")
	}
}

func TestJsonIt(t *testing.T) {
	result := JsonIt(10)
	if result != "10" {
		t.Errorf("JsonIt test failed, expect[10], got[%s]", result)
	}
	result = JsonIt(struct{ Message string }{Message: "Hello!"})
	if result != "{\"Message\":\"Hello!\"}" {
		t.Errorf("JsonIt test failed, expect[{\"Message\":\"Hello!\"}\n], got[%s]", result)
	}
}

func TestIsNativeKind(t *testing.T) {
	if !IsNativeKind(reflect.String) {
		t.Errorf("IsNativeKind test failed, expected[true], got[false]")
	}
	if IsNativeKind(reflect.Struct) {
		t.Errorf("IsNativeKind test failed, expected[false], got[true]")
	}
}

func TestGetenv(t *testing.T) {
	_ = os.Setenv("test01", "test01")
	result := Getenv("test00", "test00")
	if result != "test00" {
		t.Errorf("Getenv test failed, expected[test00], got[%s]", result)
	}
	result = Getenv("test01", "test00")
	if result != "test01" {
		t.Errorf("Getenv test failed, expected[test01], got[%s]", result)
	}
}

func TestSafeAtoi(t *testing.T) {
	result := SafeAtoi("123weq", 321)
	if result != 321 {
		t.Errorf("SafeAtoi test failed, expected[321], got[%d]", result)
	}
	result = SafeAtoi("123", 321)
	if result != 123 {
		t.Errorf("SafeAtoi test failed, expected[123], got[%d]", result)
	}
}

func TestStringPtr(t *testing.T) {
	value := StringPtr("hello")
	if value == nil {
		t.Error("StringPtr test failed nil pointer returned")
		return
	}
	if *value != "hello" {
		t.Errorf("StringPtr test failed, expected[hello], got[%s]", *value)
	}
}

func TestIsValidGUID(t *testing.T) {
	value := IsValidGUID("not a guid")
	if value == true {
		t.Error("IsValidGUID test failed, true returned for 'not a guid'")
		return
	}
	value = IsValidGUID("33967e77-ec79-4ece-9a61-55071a99279c")
	if value == false {
		t.Error("IsValidGUID test failed, false returned for '33967e77-ec79-4ece-9a61-55071a99279c'")
		return
	}
}

func TestCreateSQSUrlFromArn(t *testing.T) {
	type args struct {
		arn string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "Good",
			args: args{
				arn: "arn:aws:sqs:eu-west-1:427246389222:onboarding-state-engine",
			},
			want: "https://sqs.eu-west-1.amazonaws.com/427246389222/onboarding-state-engine",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateSQSUrlFromArn(tt.args.arn); got != tt.want {
				t.Errorf("CreateSQSUrlFromArn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateKeyValuePairs(t *testing.T) {
	type args struct {
		m map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantOne []string
	}{
		// TODO: Add test cases.
		{
			name: "Good",
			args: args{
				m: map[string]string{
					"key001":  "value",
					"key:002": "value:002",
				},
			},
			wantOne: []string{
				"key%3A002=value%3A002&key001=value",
				"key001=value&key%3A002=value%3A002",
			},
		},
		{
			name: "Empty",
			args: args{
				m: map[string]string{},
			},
			wantOne: []string{""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateKeyValuePairs(tt.args.m)
			for _, expected := range tt.wantOne {
				if got == expected {
					return
				}
			}
			t.Errorf("CreateKeyValuePairs() = %v, wanted one of %v", got, tt.wantOne)
		})
	}
}

func Test_getHeader(t *testing.T) {
	type args struct {
		headerName string
		r          *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Good",
			args: args{
				headerName: "X-Header-001",
				r: &http.Request{
					Header: map[string][]string{
						"X-Header-001": {"Value"},
					},
				},
			},
			want:    "Value",
			wantErr: false,
		},
		{
			name: "No_Header",
			args: args{
				headerName: "X-Header-002",
				r: &http.Request{
					Header: map[string][]string{
						"X-Header-001": {"Value"},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHeader(tt.args.headerName, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHeader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHeaderWithPrefixAndDecode(t *testing.T) {
	type args struct {
		headerName string
		prefix     string
		r          *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Prefix",
			args: args{
				headerName: "X-Header-001",
				r: &http.Request{
					Header: map[string][]string{
						"X-Header-001": {"Prefix YmFzZTY0LlN0ZEVuY29kaW5n"},
					},
				},
				prefix: "Prefix ",
			},
			want:    "base64.StdEncoding",
			wantErr: false,
		},
		{
			name: "No_Prefix",
			args: args{
				headerName: "X-Header-001",
				r: &http.Request{
					Header: map[string][]string{
						"X-Header-001": {"YmFzZTY0LlN0ZEVuY29kaW5n"},
					},
				},
				prefix: "Prefix ",
			},
			want:    "base64.StdEncoding",
			wantErr: false,
		},
		{
			name: "No_Header",
			args: args{
				headerName: "X-Header-002",
				r: &http.Request{
					Header: map[string][]string{
						"X-Header-001": {"YmFzZTY0LlN0ZEVuY29kaW5n"},
					},
				},
				prefix: "Prefix ",
			},
			wantErr: true,
		},
		{
			name: "Bad_B64Value",
			args: args{
				headerName: "X-Header-001",
				r: &http.Request{
					Header: map[string][]string{
						"X-Header-001": {"YmFzsdf2343sdZEVuY29kaW5n"},
					},
				},
				prefix: "Prefix ",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetHeaderWithPrefixAndDecode(tt.args.headerName, tt.args.prefix, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHeaderWithPrefixAndDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetHeaderWithPrefixAndDecode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TrimFirstRune(t *testing.T) {
	type args struct {
		s string
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "Empty_String",
			args: args{"", 1},
			want: "",
		},
		{
			name: "One_Rune_String",
			args: args{"1", 1},
			want: "",
		},
		{
			name: "_String",
			args: args{"Bye", 2},
			want: "e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimLeftChars(tt.args.s, tt.args.n); got != tt.want {
				t.Errorf("trimFirstRune() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveStatus(t *testing.T) {
	type args struct {
		r *http.Response
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"check status code 500",
			args{&http.Response{StatusCode: 500}},
			"Generic error that is encountered due to an unexpected server error.",
		},
		{
			"check status code 200",
			args{&http.Response{StatusCode: 200}},
			"The API request is successful.",
		},
		{
			"check status code empty",
			args{&http.Response{StatusCode: 0}},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveStatus(tt.args.r); got != tt.want {
				t.Errorf("exp: %#v got: %#v", tt.want, got)
			}
		})
	}
}

func TestXmlIt(t *testing.T) {
	i := 10
	result := XmlIt(i)
	if result != "10" {
		t.Errorf("exp: %v got: %v", i, result)
		return
	}
	type TestXML struct {
		Integer int
	}
	st := &TestXML{
		Integer: 10,
	}
	result = XmlIt(st)
	if result != "<TestXML><Integer>10</Integer></TestXML>" {
		t.Errorf("exp: %v got: %v", i, result)
		return
	}
}

func TestSrc(t *testing.T) {
	type args struct {
		callerDepth int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "good",
			args: args{
				callerDepth: 1,
			},
			want: "utils.TestSrc.func1 at utils_test.go:635",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Src(tt.args.callerDepth); got != tt.want {
				t.Errorf("Src() = %v, want %v", got, tt.want)
			}
		})
	}
}
