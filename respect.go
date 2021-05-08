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
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Respect compares variables a and b, recursing into their structure, and returns a list of differences,
// or nil if there are none. Some differences may not be found if an error is also returned.
//
// If a type has an Equal method, like time.Equal, it is called to check for
// equality.
func Respect(obj, respectObj interface{}) []string {
	objVal := reflect.ValueOf(obj)
	respectObjVal := reflect.ValueOf(respectObj)
	c := &cmp{
		diff:        []string{},
		buff:        []string{},
		floatFormat: fmt.Sprintf("%%.%df", FloatPrecision),
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
		return c.diff // diffs
	}
	return nil // no diffs
}

func (c *cmp) respect(objVal, respectObjVal reflect.Value, level int) {
	// Check if one value is nil, e.g. T{x: *X} and T.x is nil
	if respectObjVal.IsValid() {
		if !objVal.IsValid() {
			c.saveDiff("<nil pointer>", respectObjVal.Type())
		}
	} else {
		return
	}

	// If different types, they can't be equal
	objType := objVal.Type()
	respectObjType := respectObjVal.Type()
	if objType != respectObjType {
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
		return
	}

	// Primitive https://golang.org/pkg/reflect/#Kind
	objKind := objVal.Kind()
	respectObjKind := respectObjVal.Kind()

	// Do objVal and respectObjVal have underlying elements? Yes if they're ptr or interface.
	objElem := objKind == reflect.Ptr || objKind == reflect.Interface
	respectObjElem := respectObjKind == reflect.Ptr || respectObjKind == reflect.Interface

	// If both types implement the error interface, compare the error strings.
	// This must be done before dereferencing because the interface is on objVal
	// pointer receiver. Re https://github.com/go-test/deep/issues/31, objVal/respectObjVal might
	// be primitive kinds; see TestErrorPrimitiveKind.
	if objType.Implements(errorType) && respectObjType.Implements(errorType) {
		if (!objElem || !objVal.IsNil()) && (!respectObjElem || !respectObjVal.IsNil()) {
			aString := objVal.MethodByName("Error").Call(nil)[0].String()
			bString := respectObjVal.MethodByName("Error").Call(nil)[0].String()
			if aString != bString {
				c.saveDiff(aString, bString)
				return
			}
		}
	}

	// Dereference pointers and interface{}
	if objElem || respectObjElem {
		if objElem {
			objVal = objVal.Elem()
		}
		if respectObjElem {
			respectObjVal = respectObjVal.Elem()
		}
		c.respect(objVal, respectObjVal, level+1)
		return
	}

	switch objKind {

	/////////////////////////////////////////////////////////////////////
	// Iterable kinds
	/////////////////////////////////////////////////////////////////////

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

		aLen := objVal.Len()
		bLen := respectObjVal.Len()

		if objVal.Pointer() == respectObjVal.Pointer() && aLen == bLen {
			return
		}

		n := aLen
		if bLen > aLen {
			n = bLen
		}
		for i := 0; i < n; i++ {
			c.push(fmt.Sprintf("slice[%d]", i))
			if i < aLen && i < bLen {
				c.respect(objVal.Index(i), respectObjVal.Index(i), level+1)
			} else if i < aLen {
				c.saveDiff(objVal.Index(i), "<no value>")
			} else {
				c.saveDiff("<no value>", respectObjVal.Index(i))
			}
			c.pop()
			if len(c.diff) >= MaxDiff {
				break
			}
		}

	/////////////////////////////////////////////////////////////////////
	// Primitive kinds
	/////////////////////////////////////////////////////////////////////

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

func (c *cmp) push(name string) {
	c.buff = append(c.buff, name)
}

func (c *cmp) pop() {
	if len(c.buff) > 0 {
		c.buff = c.buff[0 : len(c.buff)-1]
	}
}

func (c *cmp) saveDiff(aval, bval interface{}) {
	if len(c.buff) > 0 {
		varName := strings.Join(c.buff, ".")
		c.diff = append(c.diff, fmt.Sprintf("%s: %v != %v", varName, aval, bval))
	} else {
		c.diff = append(c.diff, fmt.Sprintf("%v != %v", aval, bval))
	}
}
