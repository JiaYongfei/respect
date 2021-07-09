# Gomega matchers for verifying part fields' equality of a complex struct

This package provides only one matcher called respect. It's useful if you only want to check part fields' equality of a complex struct.

## Respect()

e.g.: obj.Should(Respect(respectObj, options))

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

e.g.:

Given a complex struct like below

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

Assert with common Gomega matchers. it looks like

```go
Ω(obj).ShouldNot(BeNil())
Ω(obj.Name).Should(Equal("NeZha"))
Ω(obj.Age).Should(Equal(int32(3)))
Ω(obj.Color).Should(Equal(ColorYellow))
Ω(obj.Body.Legs).Should(HaveLen(2))
Ω(obj.Body.Legs[0].Name).Should(Equal(LegRight))
Ω(obj.Body.Legs[0].Name).Should(Equal(LegLeft))
```

Assert with respect matchers, it'll be more readable

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
