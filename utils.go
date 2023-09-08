package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const (
	// ContentTypeApplicationANY const
	ContentTypeApplicationANY = "*/*"
	// ContentTypeApplicationJSON const
	ContentTypeApplicationJSON = "application/json"
	// ContentTypeApplicationXML const
	ContentTypeApplicationXML = "application/xml"
	// ContentTypeHeader const
	ContentTypeHeader = "Content-Type"
	// AuthorizationHeader const
	AuthorizationHeader = "Authorization"
	// BearerToken const
	BearerToken = "Bearer "
)

// HTTPStatusCode is a type for resolving the returned HTTP Status Code Content
type HTTPStatusCode int

// HTTPStatusCodes is a map of possible HTTP Status Code and Messages
var HTTPStatusCodes = map[HTTPStatusCode]string{
	200: "The API request is successful.",
	201: "Request fulfilled for single record insertion.",
	202: "Request fulfilled for multiple records insertion.",
	204: "There is no content available for the request.",
	304: "The requested page has not been modified. In case \"If-Modified-Since\" header is used for GET APIs",
	400: "The request or the authentication considered is invalid.",
	401: "Invalid API key provided.",
	403: "No permission to do the operation.",
	404: "Invalid request.",
	405: "The specified method is not allowed.",
	413: "The server did not accept the request while uploading a file, since the limited file size has exceeded.",
	415: "The server did not accept the request while uploading a file, since the media/ file type is not supported.",
	429: "Number of API requests per minute/day has exceeded the limit.",
	500: "Generic error that is encountered due to an unexpected server error.",
}

type GenericResponse struct {
	XMLName           *xml.Name `xml:"API3G" json:",omitempty"`
	Result            string    `xml:"Result"`
	ResultExplanation string    `xml:"ResultExplanation"`
}

func ResolveStatus(r *http.Response) string {
	if v, ok := HTTPStatusCodes[HTTPStatusCode(r.StatusCode)]; ok {
		return v
	}
	return ""
}

// GetRequestBody return array of byte from the request
func GetRequestBody(request *http.Request) ([]byte, error) {
	if request.ContentLength == 0 {
		return nil, errors.New("no body found in request")
	}
	rs, _ := ioutil.ReadAll(request.Body)
	_ = request.Body.Close()
	request.Body = ioutil.NopCloser(bytes.NewBuffer(rs))

	return rs, nil
}

// FileExists func
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ErrorToText
func ErrorToText(err GenericResponse) string {
	return fmt.Sprintf("%s;%s;;;", err.Result, err.ResultExplanation)
}

// ResponseWriterTextHtml function
func ResponseWriterTextHtml(writer http.ResponseWriter, response interface{}) {
	writer.Header().Set(ContentTypeHeader, ContentTypeApplicationXML)
	writer.Write([]byte(fmt.Sprint(response)))
}

// ResponseWriterXML function
func ResponseWriterXML(writer http.ResponseWriter, response interface{}) {
	writer.Header().Set(ContentTypeHeader, ContentTypeApplicationXML)
	xmlHeader := "<?xml version=\"1.0\" encoding=\"utf-8\"?>"
	_, _ = writer.Write([]byte(xmlHeader))
	_ = xml.NewEncoder(writer).Encode(response)
}

// ResponseWriterJSON function
func ResponseWriterJSON(writer http.ResponseWriter, response interface{}) {
	writer.Header().Set(ContentTypeHeader, ContentTypeApplicationJSON)
	//js, _ := json.Marshal(response)
	//writer.Write(js)
	_ = json.NewEncoder(writer).Encode(response)
}

// check if string slice contains a value
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// IsPtrType tester
func IsPtrType(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

// IsPtrValue tester
func IsPtrValue(t reflect.Value) bool {
	return t.Kind() == reflect.Ptr
}

// is time.Time type
func IsTimeType(t reflect.Type) bool {
	return t.String() == "time.Time"
}

// isCustomType tester
func IsCustomType(t reflect.Type) bool {
	if t.PkgPath() != "" {
		if t.PkgPath() == "syscall" {
			return false
		}
		return true
	}

	if k := t.Kind(); k == reflect.Array || k == reflect.Chan || k == reflect.Map || IsPtrType(t) || k == reflect.Slice {
		return IsCustomType(t.Elem()) || k == reflect.Map && IsCustomType(t.Key())
	} else if k == reflect.Struct {
		for i := t.NumField() - 1; i >= 0; i-- {
			if IsCustomType(t.Field(i).Type) {
				return true
			}
		}
	}
	return false
}

func Indirect(reflectValue reflect.Value) reflect.Value {
	for IsPtrValue(reflectValue) {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func IndirectType(reflectType reflect.Type) reflect.Type {
	for IsPtrType(reflectType) || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Struct:
		return false
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// ptr wraps the given value with pointer: V => *V, *V => **V, etc.
func Ptr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type()) // create a *T type.
	pv := reflect.New(pt.Elem())  // create a reflect.Value of type *T.
	pv.Elem().Set(v)              // sets pv to point to underlying value of v.
	return pv
}

func JsonIt(a interface{}) string {
	dataType := IndirectType(reflect.TypeOf(a))
	switch dataType.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Bool,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", a)
	}
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	e.Encode(a)
	return strings.TrimSuffix(buf.String(), "\n")
}

func XmlIt(a interface{}) string {
	dataType := IndirectType(reflect.TypeOf(a))
	switch dataType.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Bool,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", a)
	}
	var buf bytes.Buffer
	e := xml.NewEncoder(&buf)
	e.Encode(a)
	return strings.TrimSuffix(buf.String(), "\n")
}

func IsNativeKind(v reflect.Kind) bool {
	switch v {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Bool,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

func Getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func SafeAtoi(str string, fallback *int) *int {
	value, err := strconv.Atoi(str)
	if err != nil {
		return fallback
	}
	return &value
}

func StringPtr(str string) *string {
	return &str
}

func IsValidGUID(guid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(guid)
}

func getHeader(headerName string, r *http.Request) (string, error) {
	value := r.Header.Get(headerName)
	if len(value) == 0 {
		return "", fmt.Errorf("no %s header present", headerName)
	}
	return value, nil
}

func GetHeaderWithPrefixAndDecode(headerName string, prefix string, r *http.Request) (string, error) {
	value, err := getHeader(headerName, r)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimPrefix(value, prefix)
	bValue, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return "", err
	}
	return string(bValue), nil
}

func CreateKeyValuePairs(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=%s&", url.QueryEscape(key), url.QueryEscape(value))
	}
	return b.String()[:b.Len()-1]
}

func CreateSQSUrlFromArn(arn string) string {
	splits := strings.Split(arn, ":")
	return fmt.Sprintf("https://%s.%s.amazonaws.com/%s/%s", splits[2], splits[3], splits[4], splits[5])
}

func TrimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}

func getISO3361() map[string]string {
	return map[string]string{
		"AFGHANISTAN":                           "AF",
		"ÅLAND ISLANDS":                         "AX",
		"ALBANIA":                               "AL",
		"ALGERIA":                               "DZ",
		"AMERICAN SAMOA":                        "AS",
		"ANDORRA":                               "AD",
		"ANGOLA":                                "AO",
		"ANGUILLA":                              "AI",
		"ANTARCTICA":                            "AQ",
		"ANTIGUA AND BARBUDA":                   "AG",
		"ARGENTINA":                             "AR",
		"ARMENIA":                               "AM",
		"ARUBA":                                 "AW",
		"AUSTRALIA":                             "AU",
		"AUSTRIA":                               "AT",
		"AZERBAIJAN":                            "AZ",
		"BAHAMAS":                               "BS",
		"BAHRAIN":                               "BH",
		"BANGLADESH":                            "BD",
		"BARBADOS":                              "BB",
		"BELARUS":                               "BY",
		"BELGIUM":                               "BE",
		"BELIZE":                                "BZ",
		"BENIN":                                 "BJ",
		"BERMUDA":                               "BM",
		"BHUTAN":                                "BT",
		"BOLIVIA, PLURINATIONAL STATE OF":       "BO",
		"BONAIRE, SINT EUSTATIUS AND SABA":      "BQ",
		"BOSNIA AND HERZEGOVINA":                "BA",
		"BOTSWANA":                              "BW",
		"BOUVET ISLAND":                         "BV",
		"BRAZIL":                                "BR",
		"BRITISH INDIAN OCEAN TERRITORY":        "IO",
		"BRUNEI DARUSSALAM":                     "BN",
		"BULGARIA":                              "BG",
		"BURKINA FASO":                          "BF",
		"BURUNDI":                               "BI",
		"CAMBODIA":                              "KH",
		"CAMEROON":                              "CM",
		"CANADA":                                "CA",
		"CAPE VERDE":                            "CV",
		"CAYMAN ISLANDS":                        "KY",
		"CENTRAL AFRICAN REPUBLIC":              "CF",
		"CHAD":                                  "TD",
		"CHILE":                                 "CL",
		"CHINA":                                 "CN",
		"CHRISTMAS ISLAND":                      "CX",
		"COCOS (KEELING) ISLANDS":               "CC",
		"COLOMBIA":                              "CO",
		"COMOROS":                               "KM",
		"CONGO":                                 "CG",
		"CONGO, THE DEMOCRATIC REPUBLIC OF THE": "CD",
		"COOK ISLANDS":                          "CK",
		"COSTA RICA":                            "CR",
		"CÔTE D\"IVOIRE":                        "CI",
		"CROATIA":                               "HR",
		"CUBA":                                  "CU",
		"CURAÇAO":                               "CW",
		"CYPRUS":                                "CY",
		"CZECH REPUBLIC":                        "CZ",
		"DENMARK":                               "DK",
		"DJIBOUTI":                              "DJ",
		"DOMINICA":                              "DM",
		"DOMINICAN REPUBLIC":                    "DO",
		"ECUADOR":                               "EC",
		"EGYPT":                                 "EG",
		"EL SALVADOR":                           "SV",
		"EQUATORIAL GUINEA":                     "GQ",
		"ERITREA":                               "ER",
		"ESTONIA":                               "EE",
		"ETHIOPIA":                              "ET",
		"FALKLAND ISLANDS (MALVINAS)":           "FK",
		"FAROE ISLANDS":                         "FO",
		"FIJI":                                  "FJ",
		"FINLAND":                               "FI",
		"FRANCE":                                "FR",
		"FRENCH GUIANA":                         "GF",
		"FRENCH POLYNESIA":                      "PF",
		"FRENCH SOUTHERN TERRITORIES":           "TF",
		"GABON":                                 "GA",
		"GAMBIA":                                "GM",
		"GEORGIA":                               "GE",
		"GERMANY":                               "DE",
		"GHANA":                                 "GH",
		"GIBRALTAR":                             "GI",
		"GREECE":                                "GR",
		"GREENLAND":                             "GL",
		"GRENADA":                               "GD",
		"GUADELOUPE":                            "GP",
		"GUAM":                                  "GU",
		"GUATEMALA":                             "GT",
		"GUERNSEY":                              "GG",
		"GUINEA":                                "GN",
		"GUINEA-BISSAU":                         "GW",
		"GUYANA":                                "GY",
		"HAITI":                                 "HT",
		"HEARD ISLAND AND MCDONALD ISLANDS":     "HM",
		"HOLY SEE (VATICAN CITY STATE)":         "VA",
		"HONDURAS":                              "HN",
		"HONG KONG":                             "HK",
		"HUNGARY":                               "HU",
		"ICELAND":                               "IS",
		"INDIA":                                 "IN",
		"INDONESIA":                             "ID",
		"IRAN, ISLAMIC REPUBLIC OF":             "IR",
		"IRAQ":                                  "IQ",
		"IRELAND":                               "IE",
		"ISLE OF MAN":                           "IM",
		"ISRAEL":                                "IL",
		"ITALY":                                 "IT",
		"JAMAICA":                               "JM",
		"JAPAN":                                 "JP",
		"JERSEY":                                "JE",
		"JORDAN":                                "JO",
		"KAZAKHSTAN":                            "KZ",
		"KENYA":                                 "KE",
		"KIRIBATI":                              "KI",
		"KOREA, DEMOCRATIC PEOPLE\"S REPUBLIC OF": "KP",
		"KOREA, REPUBLIC OF":                      "KR",
		"KUWAIT":                                  "KW",
		"KYRGYZSTAN":                              "KG",
		"LAO PEOPLE\"S DEMOCRATIC REPUBLIC":       "LA",
		"LATVIA":                                  "LV",
		"LEBANON":                                 "LB",
		"LESOTHO":                                 "LS",
		"LIBERIA":                                 "LR",
		"LIBYA":                                   "LY",
		"LIECHTENSTEIN":                           "LI",
		"LITHUANIA":                               "LT",
		"LUXEMBOURG":                              "LU",
		"MACAO":                                   "MO",
		"MACEDONIA, THE FORMER YUGOSLAV REPUBLIC OF": "MK",
		"MADAGASCAR":                      "MG",
		"MALAWI":                          "MW",
		"MALAYSIA":                        "MY",
		"MALDIVES":                        "MV",
		"MALI":                            "ML",
		"MALTA":                           "MT",
		"MARSHALL ISLANDS":                "MH",
		"MARTINIQUE":                      "MQ",
		"MAURITANIA":                      "MR",
		"MAURITIUS":                       "MU",
		"MAYOTTE":                         "YT",
		"MEXICO":                          "MX",
		"MICRONESIA, FEDERATED STATES OF": "FM",
		"MOLDOVA, REPUBLIC OF":            "MD",
		"MONACO":                          "MC",
		"MONGOLIA":                        "MN",
		"MONTENEGRO":                      "ME",
		"MONTSERRAT":                      "MS",
		"MOROCCO":                         "MA",
		"MOZAMBIQUE":                      "MZ",
		"MYANMAR":                         "MM",
		"NAMIBIA":                         "NA",
		"NAURU":                           "NR",
		"NEPAL":                           "NP",
		"NETHERLANDS":                     "NL",
		"NEW CALEDONIA":                   "NC",
		"NEW ZEALAND":                     "NZ",
		"NICARAGUA":                       "NI",
		"NIGER":                           "NE",
		"NIGERIA":                         "NG",
		"NIUE":                            "NU",
		"NORFOLK ISLAND":                  "NF",
		"NORTHERN MARIANA ISLANDS":        "MP",
		"NORWAY":                          "NO",
		"OMAN":                            "OM",
		"PAKISTAN":                        "PK",
		"PALAU":                           "PW",
		"PALESTINE, STATE OF":             "PS",
		"PANAMA":                          "PA",
		"PAPUA NEW GUINEA":                "PG",
		"PARAGUAY":                        "PY",
		"PERU":                            "PE",
		"PHILIPPINES":                     "PH",
		"PITCAIRN":                        "PN",
		"POLAND":                          "PL",
		"PORTUGAL":                        "PT",
		"PUERTO RICO":                     "PR",
		"QATAR":                           "QA",
		"RÉUNION":                         "RE",
		"ROMANIA":                         "RO",
		"RUSSIAN FEDERATION":              "RU",
		"RWANDA":                          "RW",
		"SAINT BARTHÉLEMY":                "BL",
		"SAINT HELENA, ASCENSION AND TRISTAN DA CUNHA": "SH",
		"SAINT KITTS AND NEVIS":                        "KN",
		"SAINT LUCIA":                                  "LC",
		"SAINT MARTIN (FRENCH PART)":                   "MF",
		"SAINT PIERRE AND MIQUELON":                    "PM",
		"SAINT VINCENT AND THE GRENADINES":             "VC",
		"SAMOA":                                        "WS",
		"SAN MARINO":                                   "SM",
		"SAO TOME AND PRINCIPE":                        "ST",
		"SAUDI ARABIA":                                 "SA",
		"SENEGAL":                                      "SN",
		"SERBIA":                                       "RS",
		"SEYCHELLES":                                   "SC",
		"SIERRA LEONE":                                 "SL",
		"SINGAPORE":                                    "SG",
		"SINT MAARTEN (DUTCH PART)":                    "SX",
		"SLOVAKIA":                                     "SK",
		"SLOVENIA":                                     "SI",
		"SOLOMON ISLANDS":                              "SB",
		"SOMALIA":                                      "SO",
		"SOUTH AFRICA":                                 "ZA",
		"SOUTH GEORGIA AND THE SOUTH SANDWICH ISLANDS": "GS",
		"SOUTH SUDAN":                          "SS",
		"SPAIN":                                "ES",
		"SRI LANKA":                            "LK",
		"SUDAN":                                "SD",
		"SURINAME":                             "SR",
		"SVALBARD AND JAN MAYEN":               "SJ",
		"SWAZILAND":                            "SZ",
		"SWEDEN":                               "SE",
		"SWITZERLAND":                          "CH",
		"SYRIAN ARAB REPUBLIC":                 "SY",
		"TAIWAN, PROVINCE OF CHINA":            "TW",
		"TAJIKISTAN":                           "TJ",
		"TANZANIA, UNITED REPUBLIC OF":         "TZ",
		"THAILAND":                             "TH",
		"TIMOR-LESTE":                          "TL",
		"TOGO":                                 "TG",
		"TOKELAU":                              "TK",
		"TONGA":                                "TO",
		"TRINIDAD AND TOBAGO":                  "TT",
		"TUNISIA":                              "TN",
		"TURKEY":                               "TR",
		"TURKMENISTAN":                         "TM",
		"TURKS AND CAICOS ISLANDS":             "TC",
		"TUVALU":                               "TV",
		"UGANDA":                               "UG",
		"UKRAINE":                              "UA",
		"UNITED ARAB EMIRATES":                 "AE",
		"UNITED KINGDOM":                       "GB",
		"UNITED STATES":                        "US",
		"UNITED STATES MINOR OUTLYING ISLANDS": "UM",
		"URUGUAY":                              "UY",
		"UZBEKISTAN":                           "UZ",
		"VANUATU":                              "VU",
		"VENEZUELA, BOLIVARIAN REPUBLIC OF":    "VE",
		"VIET NAM":                             "VN",
		"VIRGIN ISLANDS, BRITISH":              "VG",
		"VIRGIN ISLANDS, U.S.":                 "VI",
		"WALLIS AND FUTUNA":                    "WF",
		"WESTERN SAHARA":                       "EH",
		"YEMEN":                                "YE",
		"ZAMBIA":                               "ZM",
		"ZIMBABWE":                             "ZW",
	}
}

func GetISOCodeCountry(country string) string {
	iso := getISO3361()
	return iso[strings.ToUpper(country)]
}

func Src(callerDepth int) string {
	// Determine caller func
	pc, file, lineno, ok := runtime.Caller(callerDepth)
	src := ""
	if ok {
		slice := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		src = slice[len(slice)-1]
		slice = strings.Split(file, "/")
		file := slice[len(slice)-1]
		src = fmt.Sprintf("%s at %s:%d", src, file, lineno)
	}
	return src
}
