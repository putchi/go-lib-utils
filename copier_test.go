package utils

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type User struct {
	Name     string
	Birthday *time.Time
	Nickname string
	Role     string
	Age      int32
	FakeAge  *int32
	Notes    []string
	flags    []byte
}

func (user User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	Name      string
	Birthday  *time.Time
	Nickname  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	flags     []byte
}

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}

func checkEmployee(employee Employee, user User, t *testing.T, testCase string) {
	if employee.Name != user.Name {
		t.Errorf("%v: Name haven't been copied correctly.", testCase)
	}
	if employee.Nickname == nil || *employee.Nickname != user.Nickname {
		t.Errorf("%v: NickName haven't been copied correctly.", testCase)
	}
	if employee.Birthday == nil && user.Birthday != nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday == nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday != nil &&
		!employee.Birthday.Equal(*(user.Birthday)) {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Age != int64(user.Age) {
		t.Errorf("%v: Age haven't been copied correctly.", testCase)
	}
	if user.FakeAge != nil && employee.FakeAge != int(*user.FakeAge) {
		t.Errorf("%v: FakeAge haven't been copied correctly.", testCase)
	}
	if employee.DoubleAge != user.DoubleAge() {
		t.Errorf("%v: Copy from method doesn't work", testCase)
	}
	if employee.SuperRule != "Super "+user.Role {
		t.Errorf("%v: Copy to method doesn't work", testCase)
	}
	if !reflect.DeepEqual(employee.Notes, user.Notes) {
		t.Errorf("%v: Copy from slice doen't work", testCase)
	}
}

func TestCopySameStructWithPointerField(t *testing.T) {
	var fakeAge int32 = 12
	var currentTime = time.Now()
	user := &User{Birthday: &currentTime, Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	newUser := &User{}
	_ = Copy(newUser, user)
	if user.Birthday == newUser.Birthday {
		t.Errorf("TestCopySameStructWithPointerField: copy Birthday failed since they need to have different address")
	}

	if user.FakeAge == newUser.FakeAge {
		t.Errorf("TestCopySameStructWithPointerField: copy FakeAge failed since they need to have different address")
	}
}

func checkEmployee2(employee Employee, user *User, t *testing.T, testCase string) {
	if user == nil {
		if employee.Name != "" || employee.Nickname != nil || employee.Birthday != nil || employee.Age != 0 ||
			employee.DoubleAge != 0 || employee.FakeAge != 0 || employee.SuperRule != "" || employee.Notes != nil {
			t.Errorf("%v : employee should be empty", testCase)
		}
		return
	}

	checkEmployee(employee, *user, t, testCase)
}

func TestCopyStruct(t *testing.T) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	employee := Employee{}

	if err := Copy(employee, &user); err == nil {
		t.Errorf("Copy to unaddressable value should get error")
	}

	_ = Copy(&employee, &user)
	checkEmployee(employee, user, t, "Copy From Ptr To Ptr")

	employee2 := Employee{}
	_ = Copy(&employee2, user)
	checkEmployee(employee2, user, t, "Copy From Struct To Ptr")

	employee3 := Employee{}
	ptrToUser := &user
	_ = Copy(&employee3, &ptrToUser)
	checkEmployee(employee3, user, t, "Copy From Double Ptr To Ptr")

	employee4 := &Employee{}
	_ = Copy(&employee4, user)
	checkEmployee(*employee4, user, t, "Copy From Ptr To Double Ptr")
}

func TestCopyFromStructToSlice(t *testing.T) {
	user := User{Name: "Jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}
	var employees []Employee

	if err := Copy(employees, &user); err != nil && len(employees) != 0 {
		t.Errorf("Copy to unaddressable value should get error")
	}

	if _ = Copy(&employees, &user); len(employees) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(employees[0], user, t, "Copy From Struct To Slice Ptr")
	}

	employees2 := &[]Employee{}
	if _ = Copy(&employees2, user); len(*employees2) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee((*employees2)[0], user, t, "Copy From Struct To Double Slice Ptr")
	}

	var employees3 []*Employee
	if _ = Copy(&employees3, user); len(employees3) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*(employees3[0]), user, t, "Copy From Struct To Ptr Slice Ptr")
	}

	employees4 := &[]*Employee{}
	if _ = Copy(&employees4, user); len(*employees4) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*((*employees4)[0]), user, t, "Copy From Struct To Double Ptr Slice Ptr")
	}
}

func TestCopyFromSliceToSlice(t *testing.T) {
	users := []User{{Name: "Jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}, {Name: "Jinzhu2", Age: 22, Role: "Dev", Notes: []string{"hello world", "hello"}}}
	var employees []Employee

	if _ = Copy(&employees, users); len(employees) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(employees[0], users[0], t, "Copy From Slice To Slice Ptr @ 1")
		checkEmployee(employees[1], users[1], t, "Copy From Slice To Slice Ptr @ 2")
	}

	employees2 := &[]Employee{}
	if _ = Copy(&employees2, &users); len(*employees2) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee((*employees2)[0], users[0], t, "Copy From Slice Ptr To Double Slice Ptr @ 1")
		checkEmployee((*employees2)[1], users[1], t, "Copy From Slice Ptr To Double Slice Ptr @ 2")
	}

	var employees3 []*Employee
	if _ = Copy(&employees3, users); len(employees3) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(*(employees3[0]), users[0], t, "Copy From Slice To Ptr Slice Ptr @ 1")
		checkEmployee(*(employees3[1]), users[1], t, "Copy From Slice To Ptr Slice Ptr @ 2")
	}

	employees4 := &[]*Employee{}
	if _ = Copy(&employees4, users); len(*employees4) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(*((*employees4)[0]), users[0], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 1")
		checkEmployee(*((*employees4)[1]), users[1], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 2")
	}
}

func TestCopyFromSliceToSlice2(t *testing.T) {
	users := []*User{{Name: "Jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}, nil}
	var employees []Employee

	if _ = Copy(&employees, users); len(employees) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2(employees[0], users[0], t, "Copy From Slice To Slice Ptr @ 1")
		checkEmployee2(employees[1], users[1], t, "Copy From Slice To Slice Ptr @ 2")
	}

	employees2 := &[]Employee{}
	if _ = Copy(&employees2, &users); len(*employees2) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2((*employees2)[0], users[0], t, "Copy From Slice Ptr To Double Slice Ptr @ 1")
		checkEmployee2((*employees2)[1], users[1], t, "Copy From Slice Ptr To Double Slice Ptr @ 2")
	}

	var employees3 []*Employee
	if _ = Copy(&employees3, users); len(employees3) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2(*(employees3[0]), users[0], t, "Copy From Slice To Ptr Slice Ptr @ 1")
		checkEmployee2(*(employees3[1]), users[1], t, "Copy From Slice To Ptr Slice Ptr @ 2")
	}

	employees4 := &[]*Employee{}
	if _ = Copy(&employees4, users); len(*employees4) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2(*((*employees4)[0]), users[0], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 1")
		checkEmployee2(*((*employees4)[1]), users[1], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 2")
	}
}

func TestEmbeddedAndBase(t *testing.T) {
	type Base struct {
		BaseField1 int
		BaseField2 int
		User       *User
	}

	type Embed struct {
		EmbedField1 int
		EmbedField2 int
		Base
	}

	base := Base{}
	embeded := Embed{}
	embeded.BaseField1 = 1
	embeded.BaseField2 = 2
	embeded.EmbedField1 = 3
	embeded.EmbedField2 = 4

	user := User{
		Name: "testName",
	}
	embeded.User = &user

	_ = Copy(&base, &embeded)

	if base.BaseField1 != 1 || base.User.Name != "testName" {
		t.Error("Embedded fields not copied")
	}

	base.BaseField1 = 11
	base.BaseField2 = 12
	user1 := User{
		Name: "testName1",
	}
	base.User = &user1

	_ = Copy(&embeded, &base)

	if embeded.BaseField1 != 11 || embeded.User.Name != "testName1" {
		t.Error("base fields not copied")
	}
}

type structSameName1 struct {
	A string
	B int64
	C time.Time
}

type structSameName2 struct {
	A string
	B time.Time
	C int64
}

func TestCopyFieldsWithSameNameButDifferentTypes(t *testing.T) {
	obj1 := structSameName1{A: "123", B: 2, C: time.Now()}
	obj2 := &structSameName2{}
	err := Copy(obj2, &obj1)
	if err != nil {
		t.Error("Should not raise error")
	}

	if obj2.A != obj1.A {
		t.Errorf("Field A should be copied")
	}
}

type ScannerValue struct {
	V int
}

func (s *ScannerValue) Scan(_ interface{}) error {
	return errors.New("I failed")
}

type ScannerStruct struct {
	V *ScannerValue
}

type ScannerStructTo struct {
	V *ScannerValue
}

func TestScanner(t *testing.T) {
	s := &ScannerStruct{
		V: &ScannerValue{
			V: 12,
		},
	}

	s2 := &ScannerStructTo{}

	err := Copy(s2, s)
	if err != nil {
		t.Error("Should not raise error")
	}

	if s.V.V != s2.V.V {
		t.Errorf("Field V should be copied")
	}
}

type TypeStruct1 struct {
	Field1 string
	Field2 string
	Field3 TypeStruct2
	Field4 *TypeStruct2
	Field5 []*TypeStruct2
	Field6 []TypeStruct2
	Field7 []*TypeStruct2
	Field8 []TypeStruct2
}

type TypeStruct2 struct {
	Field1 int
	Field2 string
}

type TypeStruct3 struct {
	Field1 interface{}
	Field2 string
	Field3 TypeStruct4
	Field4 *TypeStruct4
	Field5 []*TypeStruct4
	Field6 []*TypeStruct4
	Field7 []TypeStruct4
	Field8 []TypeStruct4
}

type TypeStruct4 struct {
	field1 int
	Field2 string
}

func (t *TypeStruct4) Field1(i int) {
	t.field1 = i
}

func TestCopyDifferentFieldType(t *testing.T) {
	ts := &TypeStruct1{
		Field1: "str1",
		Field2: "str2",
	}
	ts2 := &TypeStruct2{}

	_ = Copy(ts2, ts)

	if ts2.Field2 != ts.Field2 || ts2.Field1 != 0 {
		t.Errorf("Should be able to copy from ts to ts2")
	}
}

func TestCopyDifferentTypeMethod(t *testing.T) {
	ts := &TypeStruct1{
		Field1: "str1",
		Field2: "str2",
	}
	ts4 := &TypeStruct4{}

	_ = Copy(ts4, ts)

	if ts4.Field2 != ts.Field2 || ts4.field1 != 0 {
		t.Errorf("Should be able to copy from ts to ts4")
	}
}

func TestAssignableType(t *testing.T) {
	ts := &TypeStruct1{
		Field1: "str1",
		Field2: "str2",
		Field3: TypeStruct2{
			Field1: 666,
			Field2: "str2",
		},
		Field4: &TypeStruct2{
			Field1: 666,
			Field2: "str2",
		},
		Field5: []*TypeStruct2{
			{
				Field1: 666,
				Field2: "str2",
			},
		},
		Field6: []TypeStruct2{
			{
				Field1: 666,
				Field2: "str2",
			},
		},
		Field7: []*TypeStruct2{
			{
				Field1: 666,
				Field2: "str2",
			},
		},
	}

	ts3 := &TypeStruct3{}

	_ = Copy(&ts3, &ts)

	if v, ok := ts3.Field1.(string); !ok {
		t.Error("Assign to interface{} type did not succeed")
	} else if v != "str1" {
		t.Error("String haven't been copied correctly")
	}

	if ts3.Field2 != ts.Field2 {
		t.Errorf("Field2 should be copied")
	}

	checkType2WithType4(ts.Field3, ts3.Field3, t, "Field3")
	checkType2WithType4(*ts.Field4, *ts3.Field4, t, "Field4")

	for idx, f := range ts.Field5 {
		checkType2WithType4(*f, *(ts3.Field5[idx]), t, "Field5")
	}

	for idx, f := range ts.Field6 {
		checkType2WithType4(f, *(ts3.Field6[idx]), t, "Field6")
	}

	for idx, f := range ts.Field7 {
		checkType2WithType4(*f, ts3.Field7[idx], t, "Field7")
	}

	for idx, f := range ts.Field8 {
		checkType2WithType4(f, ts3.Field8[idx], t, "Field8")
	}
}

func checkType2WithType4(t2 TypeStruct2, t4 TypeStruct4, t *testing.T, testCase string) {
	if t2.Field1 != t4.field1 || t2.Field2 != t4.Field2 {
		t.Errorf("%v: type struct 4 and type struct 2 is not equal", testCase)
	}
}

func TestCopyFromNotValid(t *testing.T) {
	type A struct{}
	dest := &A{}

	_ = Copy(dest, nil)
	if dest == nil {
		t.Error("Copy test for nil src failed")
	}
}

func TestCopySimpleSet(t *testing.T) {
	src := StringPtr("10")
	dest := StringPtr("5")
	_ = Copy(dest, src)
	if *dest != *src {
		t.Errorf("Copy test failed, expected[%s], got[%s]", *src, *dest)
	}
}
