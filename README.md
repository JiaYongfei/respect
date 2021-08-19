# Gomega matchers for verifying part fields' equality of a complex struct

This package provides only one matcher called respect. It's useful if you only want to check part fields' equality of a complex struct.

## Respect()

e.g.: `obj.Should(Respect(respectObj, options))`

Respect means:
1. If obj and respectObj are primitive types, they should be equal with each other.
2. If obj and respectObj are slice/array type, obj should be a superset of respectObj and elements in obj should respect the corresponding elements in respectObj.

   If the kind of respectOjb items is reflect.Struct, below is the way we use to find the corresponding elements.
   
   Use all the valid/non-zero string/*string fields of respectObj items as the identifier to find the corresponding element in obj.
   
   If `LengthMatters` option provided, they should have same length. If `OrderMatters` option provided, they'll be compared one by one in order.
   
3. If obj and respectObj are map type, obj should contain all the key value pair in respectObj.
4. If obj and respectObj are struct type, obj should respect all the field value (except zero values field) in respectObj.
   
   If `ZeroValueMatters` option provided, zero values in respectObj should also be respected.
   
   Be careful with the non-pointer field in respectObj struct, these field will be considered as zero value if omitted and participate into the comparison if `ZeroValueMatters` option provided. 


## Given a complex struct like below

```go
complexObj = &Person{
	Name:        "NeZha",
	Age:         3,
	description: "I'm NeZha!!!",
	Color:       ColorYellow,
	Body: Body{
		Head: &Head{
			Mouth: &mouthBig,
			Eyes: []*string{
				&EyeBig,
				&EyeBig,
			},
		},
		Arms: []string{"left", "right"},
		Legs: []*Leg{
			{
				Name: &LegLeft,
			},
			{
				Name: &LegRight,
			},
		},
	},
	Memory: map[string]string{
		"1+1": "2",
		"2*2": "4",
	},
}
```

### Assert with common Gomega matchers. 

1. Lots of code to write
2. The left assertions won't execute if one of the previous assertions failed
3. No complete obj info provided if assertion failed.

```go
Ω(obj).ShouldNot(BeNil())
Ω(obj.Name).Should(Equal("NeZha"))
Ω(obj.Age).Should(Equal(int32(3)))
Ω(obj.Color).Should(Equal(ColorYellow))
Ω(obj.Body.Legs).Should(HaveLen(2))
Ω(obj.Body.Legs[0].Name).Should(Equal(LegRight))
Ω(obj.Body.Legs[0].Name).Should(Equal(LegLeft))
```

### Assert with gstruct.

1. Lots of code to write and a little bit complicated
2. Field name are write in string which is error-prone

```go
idFn := func(index int, _ interface{}) string {
	return strconv.Itoa(index)
}
Ω(obj).Should(PointTo(MatchFields(IgnoreExtras, Fields{
	"Name": Equal("NeZha"),
	"Age": Equal(int32(3)),
	"Color": Equal(ColorYellow),
	"Body":MatchFields(IgnoreExtras,Fields{
		"Legs":MatchElementsWithIndex(idFn, IgnoreExtras, Elements{
			"0": PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": PointTo(Equal(LegLeft)),
			})),
			"1": PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": PointTo(Equal(LegRight)),
			})),
		}),
	}),
})))
```

failure info if assertion failed

```go
Expected
  <string>: Person
to match fields: {
.Name:
Expected
    <string>: NeZha
to equal
    <string>: AoBing
.Age:
Expected
    <int32>: 3
to equal
    <int32>: 4
.Body.Legs:
unexpected element 1
}
```

### Assert with respect matcher, it'll be more readable

```go
Ω(complexObj).Should(Respect(&Person{
	Name:  "NeZha",
	Age:   int32(3),
	Color: ColorYellow,
	Body: Body{
		Legs: []*Leg{
		    {
			    Name: &LegRight,
		    },
		    {
			    Name: &LegLeft,
		    },
		},
	},
}, respect.LengthMatters))
```

What's more, all the detail information (obj, respectObj and the disrespect parts) will be print out if assertion failed which makes the analysis easier.

```go
Expected
    <*test_test.Person | 0xc00020e000>: {
        Name: "NeZha",
        Age: 3,
        Body: {
            Head: {
                Mouth: "Big Mouth",
                Eyes: ["Big Eye", "Big Eye"],
            },
            Arms: ["left", "right"],
        }
    }
to respect
    <*test_test.Person | 0xc00020e070>: {
        Name: "AoBing",
        Age: 4,
        Body: {Head: nil, Arms: ["left"], Legs: nil},
    }
Disrespect parts are:
Name: NeZha != AoBing
Age: 3 != 4
Body.Arms.<len>: 2 > 1
```
