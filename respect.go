package respect

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// MaxDiff specifies the maximum number of differences to return.
	MaxDiff = 10

	// FloatPrecision is the number of decimal places to round float values
	// to when comparing.
	FloatPrecision = 10
)

type cmp struct {
	diff        []string
	buff        []string
	floatFormat string
	options     Options
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Respect check if obj respect the respectObj by recursing into their structure, and returns a list of differences,
// or nil if there are none.
//
// Respect means:
// 1. if obj and respectObj are primitive types, they should be equal with each other.
// 2. if obj and respectObj are slice/array type, obj should be a superset of respectObj and elements in obj should
//    respect the corresponding elements in respectObj. If the slice/array items' kind is reflect.Struct, below is the
//    way we used to find the corresponding elements.
//    Use all the valid/non-zero string/*string fields of respectObj as the identifier to find the corresponding element
//    in obj.
//    If LengthMatters option provided, they should have same length. If OrderMatters option provided, they'll
//    be compared one by one in order.
// 3. if obj and respectObj are map type, obj should contain all the key value pair in respectObj.
// 4. if obj and respectObj are struct type, obj should contains all the fields and respect their value in respectObj.
//    Reminder: Be care of the non-pointer field in respectObj struct, these field will be considered as zero value if
//    omitted and participate into the comparison which might lead to unexpected result
func Respect(obj, respectObj interface{}, respectOptions ...Options) []string {
	objVal := reflect.ValueOf(obj)
	respectObjVal := reflect.ValueOf(respectObj)

	var options Options
	for _, option := range respectOptions {
		options = options | option
	}
	c := &cmp{
		diff:        []string{},
		buff:        []string{},
		floatFormat: fmt.Sprintf("%%.%df", FloatPrecision),
		options:     options,
	}

	if obj == nil && respectObj == nil {
		return nil
	} else if obj == nil && respectObj != nil {
		c.saveDiff("<nil pointer>", respectObj)
	} else if obj != nil && respectObj == nil {
		c.saveDiff(obj, "<nil pointer>")
	}
	if len(c.diff) > 0 {
		return c.diff
	}

	c.respect(objVal, respectObjVal, 0)
	if len(c.diff) > 0 {
		return c.diff
	}
	return nil
}

func (c *cmp) respect(objVal, respectObjVal reflect.Value, level int) {
	// Check if one value is nil, e.g. T{x: *X} and T.x is nil
	if !respectObjVal.IsValid() {
		return
	}

	if !objVal.IsValid() {
		c.saveDiff("<nil pointer>", respectObjVal.Type())
		return
	}

	// If different types, they can't be equal
	objType := objVal.Type()
	respectObjType := respectObjVal.Type()
	if objType != respectObjType {
		c.push("<type>")
		// Built-in types don't have objVal name, so don't report [3]int != [2]int as " != "
		if respectObjType.Name() == "" || respectObjType.Name() != objType.Name() {
			c.saveDiff(objType, respectObjType)
		} else {
			// Type names can be the same, e.g. pkg/v1.Error and pkg/v2.Error
			// are both exported as pkg, so unless we include the full pkg path
			// the diff will be "pkg.Error != pkg.Error"
			// https://github.com/go-test/deep/issues/39
			aFullType := objType.PkgPath() + "." + objType.Name()
			bFullType := respectObjType.PkgPath() + "." + respectObjType.Name()
			c.saveDiff(aFullType, bFullType)
		}
		c.pop()
		return
	}

	// Primitive https://golang.org/pkg/reflect/#Kind

	// If both types implement the error interface, compare the error strings.
	// This must be done before dereferencing because the interface is on objVal
	// pointer receiver. Re https://github.com/go-test/deep/issues/31, objVal/respectObjVal might
	// be primitive kinds; see TestErrorPrimitiveKind.
	//if objType.Implements(errorType) && respectObjType.Implements(errorType) {
	//	if (!objElem || !objVal.IsNil()) && (!respectObjElem || !respectObjVal.IsNil()) {
	//		aString := objVal.MethodByName("Error").Call(nil)[0].String()
	//		bString := respectObjVal.MethodByName("Error").Call(nil)[0].String()
	//		if aString != bString {
	//			c.saveDiff(aString, bString)
	//			return
	//		}
	//	}
	//}

	// Ignore the zero values if ZeroValueMatters option not set
	if c.options&ZeroValueMatters == 0 && respectObjVal.IsZero() {
		return
	}

	switch respectObjVal.Kind() {
	case reflect.Struct:
		/*
			The variables are structs like:
				type T struct {
					FirstName string
					LastName  string
				}
			Type = <pkg>.T, Kind = reflect.Struct
			Iterate through the fields (FirstName, LastName), recurse into their values.
		*/

		// Types with an Equal() method, like time.Time, only if struct field
		// is exported (CanInterface)
		if eqFunc := objVal.MethodByName("Equal"); eqFunc.IsValid() && eqFunc.CanInterface() {
			// Handle https://github.com/go-test/deep/issues/15:
			// Don't call T.Equal if the method is from an embedded struct, like:
			//   type Foo struct { time.Time }
			// First, we'll encounter Equal(Ttime, time.Time) but if we pass respectObjVal
			// as the 2nd arg we'll panic: "Call using pkg.Foo as type time.Time"
			// As far as I can tell, there's no way to see that the method is from
			// time.Time not Foo. So we check the type of the 1st (0) arg and skip
			// unless it's respectObjVal type. Later, we'll encounter the time.Time anonymous/
			// embedded field and then we'll have Equal(time.Time, time.Time).
			funcType := eqFunc.Type()
			if funcType.NumIn() == 1 && funcType.In(0) == respectObjType {
				retVals := eqFunc.Call([]reflect.Value{respectObjVal})
				if !retVals[0].Bool() {
					c.saveDiff(objVal, respectObjVal)
				}
				return
			}
		}

		for i := 0; i < respectObjVal.NumField(); i++ {
			if respectObjType.Field(i).PkgPath != "" {
				continue // skip unexported field, e.g. s in type T struct {s string}
			}

			fieldName := respectObjType.Field(i).Name
			c.push(fieldName) // push field name to buff

			// Get the Value for each field, e.g. FirstName has Type = string,
			// Kind = reflect.String.
			objF := objVal.FieldByName(fieldName)
			respectObjF := respectObjVal.Field(i)

			// Recurse to compare the field values
			c.respect(objF, respectObjF, level+1)

			c.pop() // pop field name from buff

			if len(c.diff) >= MaxDiff {
				break
			}
		}
	case reflect.Map:
		/*
			The variables are maps like:
				map[string]int{
					"foo": 1,
					"bar": 2,
				}
			Type = map[string]int, Kind = reflect.Map
			Or:
				type T map[string]int{}
			Type = <pkg>.T, Kind = reflect.Map
			Iterate through the map keys (foo, bar), recurse into their values.
		*/
		if !respectObjVal.IsNil() && respectObjVal.Len() != 0 {
			if objVal.IsNil() {
				c.saveDiff("<nil map>", respectObjVal)
			}
		} else {
			return
		}

		if objVal.Pointer() == respectObjVal.Pointer() {
			return
		}

		for _, key := range respectObjVal.MapKeys() {
			c.push(fmt.Sprintf("map[%v]", key))

			aVal := objVal.MapIndex(key)
			bVal := respectObjVal.MapIndex(key)
			if aVal.IsValid() {
				c.respect(aVal, bVal, level+1)
			} else {
				c.saveDiff("<does not have key>", bVal)
			}

			c.pop()

			if len(c.diff) >= MaxDiff {
				return
			}
		}
	case reflect.Array:
		n := respectObjVal.Len()
		for i := 0; i < n; i++ {
			c.push(fmt.Sprintf("array[%d]", i))
			c.respect(objVal.Index(i), respectObjVal.Index(i), level+1)
			c.pop()
			if len(c.diff) >= MaxDiff {
				break
			}
		}
	case reflect.Slice:
		if !respectObjVal.IsNil() && respectObjVal.Len() != 0 {
			if objVal.IsNil() {
				c.saveDiff("<nil slice>", respectObjVal)
			}
		} else {
			return
		}

		objLen := objVal.Len()
		respectObjLen := respectObjVal.Len()

		if objLen == respectObjLen {
			if objVal.Pointer() == respectObjVal.Pointer() {
				return
			}
			if respectObjLen == 0 {
				return
			}
		} else if objLen < respectObjLen {
			c.push("<len>")
			c.saveDiff_(objLen, respectObjLen, "<")
			c.pop()
			return
		} else if c.options&LengthMatters != 0 {
			c.push("<len>")
			c.saveDiff_(objLen, respectObjLen, ">")
			c.pop()
		}

		if c.options&OrderMatters != 0 || respectObjLen <= 1 && objLen == 1 {
			// compared one by one
			for i := 0; i < respectObjLen; i++ {
				c.push(fmt.Sprintf("[%v]", i))
				c.respect(objVal.Index(i), respectObjVal.Index(i), level+1)
				c.pop()
				if len(c.diff) >= MaxDiff {
					break
				}
			}
		} else {
			c.respectSliceIgnoreOrder(objVal, respectObjVal, level)
		}
	case reflect.Ptr, reflect.Interface:
		// Do objVal and respectObjVal have underlying elements? Yes if they're ptr or interface.
		// Dereference pointers and interface{}
		objVal = objVal.Elem()
		respectObjVal = respectObjVal.Elem()
		c.respect(objVal, respectObjVal, level+1)
	case reflect.Float32, reflect.Float64:
		// Round floats to FloatPrecision decimal places to compare with
		// user-defined precision. As is commonly know, floats have "imprecision"
		// such that 0.1 becomes 0.100000001490116119384765625. This cannot
		// be avoided; it can only be handled. Issue 30 suggested that floats
		// be compared using an epsilon: equal = |objVal-respectObjVal| < epsilon.
		// In many cases the result is the same, but I think epsilon is objVal little
		// less clear for users to reason about. See issue 30 for details.
		aval := fmt.Sprintf(c.floatFormat, objVal.Float())
		bval := fmt.Sprintf(c.floatFormat, respectObjVal.Float())
		if aval != bval {
			c.saveDiff(objVal.Float(), respectObjVal.Float())
		}
	case reflect.Bool:
		if objVal.Bool() != respectObjVal.Bool() {
			c.saveDiff(objVal.Bool(), respectObjVal.Bool())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if objVal.Int() != respectObjVal.Int() {
			c.saveDiff(objVal.Int(), respectObjVal.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if objVal.Uint() != respectObjVal.Uint() {
			c.saveDiff(objVal.Uint(), respectObjVal.Uint())
		}
	case reflect.String:
		if objVal.String() != respectObjVal.String() {
			c.saveDiff(objVal.String(), respectObjVal.String())
		}
	}
}

// Check if slice objVal respect slice respectObjVal without considering the items order
func (c *cmp) respectSliceIgnoreOrder(objVal, respectObjVal reflect.Value, level int) {
	// check slice items' kind. Dereference it if it's interface or pointer
	itemKind := valueType(respectObjVal.Index(0)).Kind()
	switch itemKind {
	case reflect.Struct:
		respectObjItemVal := valueType(respectObjVal.Index(0))
		// Use all the valid string/*string field as the identifier
		var fieldNames []string
		for i := 0; i < respectObjItemVal.NumField(); i++ {
			if respectObjItemVal.Field(i).IsValid() &&
				!respectObjItemVal.Field(i).IsZero() &&
				valueType(respectObjItemVal.Field(i)).Kind() == reflect.String {
				fieldNames = append(fieldNames, respectObjItemVal.Type().Field(i).Name)
			}
		}
		if len(fieldNames) == 0 {
			c.save("<non valid field identifier was found>")
			return
		}
		for i := 0; i < respectObjVal.Len(); i++ {
			c.push(fmt.Sprintf("[%v]", i))
			respectObjItemVal := valueType(respectObjVal.Index(i))
			respectHash := structHash(respectObjItemVal, fieldNames)
			found := false
			for j := 0; j < objVal.Len(); j++ {
				objItemVal := valueType(objVal.Index(j))
				if structHash(objItemVal, fieldNames) == respectHash {
					found = true
					c.respect(objVal.Index(j), respectObjVal.Index(i), level+1)
					break
				}
			}
			if !found {
				c.push(strings.Join(fieldNames, "-"))
				c.saveDiff("<not found>", respectHash)
				c.pop()
			}
			if len(c.diff) >= MaxDiff {
				break
			}
			c.pop()
		}
	case reflect.String:
		// contains all
		var dirtyObjIndex []int
		for i := 0; i < respectObjVal.Len(); i++ {
			var found bool
			for j := 0; j < objVal.Len(); j++ {
				if contains(dirtyObjIndex, j) {
					continue
				}
				if objVal.Index(j).String() == respectObjVal.Index(i).String() {
					found = true
					dirtyObjIndex = append(dirtyObjIndex, j)
					break
				}
			}
			if found {
				continue
			} else {
				c.push("item")
				c.saveDiff("<not found>", respectObjVal.Index(i).String())
				c.pop()
			}
		}
	}
}

func structHash(v reflect.Value, fieldNames []string) string {
	var respectHash []string
	for _, fn := range fieldNames {
		respectHash = append(respectHash, valueType(v.FieldByName(fn)).String())
	}
	return strings.Join(respectHash, "-")
}

func valueType(v reflect.Value) reflect.Value {
	if needDeref(v) {
		return v.Elem()
	}
	return v
}

func needDeref(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface
}

func (c *cmp) push(name string) {
	c.buff = append(c.buff, name)
}

func (c *cmp) pop() {
	if len(c.buff) > 0 {
		c.buff = c.buff[0 : len(c.buff)-1]
	}
}

func (c *cmp) saveDiff(aval, bval interface{}) {
	c.saveDiff_(aval, bval, "!=")
}

func (c *cmp) saveDiff_(aval, bval interface{}, operator string) {
	c.save(fmt.Sprintf("%v %v %v", aval, operator, bval))
}

func (c *cmp) save(msg string) {
	if len(c.buff) > 0 {
		varName := strings.Join(c.buff, ".")
		c.diff = append(c.diff, fmt.Sprintf("%s: %v", varName, msg))
	} else {
		c.diff = append(c.diff, msg)
	}
}
