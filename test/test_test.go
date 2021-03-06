package test_test

import (
	"github.com/JiaYongfei/respect"
	. "github.com/JiaYongfei/respect/gomega"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Color string

const (
	ColorYellow Color = "Yellow"
	ColorBlack  Color = "Black"
	ColorWhite  Color = "White"
)

type Head struct {
	Mouth *string
	Eyes  []*string
}

type Leg struct {
	Name      *string
	LongShort string
}

type Body struct {
	Head *Head
	Arms []string
	Legs []*Leg
}

type Person struct {
	Name        string
	Age         int32
	Description *string
	Color       Color
	Body        Body
	Memory      map[string]string
}

var (
	obj *Person

	mouthBig      = "Big Mouth"
	mouthSmall    = "Small Mouth"
	EyeBig        = "Big Eye"
	EyeSmall      = "Small Eye"
	LegLeft       = "Left Leg"
	LegRight      = "Right Leg"
	LegAdditional = "Additional Leg"
	Description   = "I'm NeZha!!!"
)

var _ = BeforeSuite(func() {
	obj = &Person{
		Name:        "NeZha",
		Age:         3,
		Description: &Description,
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
})

var _ = Describe("Test", func() {

	It("Should have same type", func() {
		Ω(obj).ShouldNot(Respect(Person{ // Value type
			Name:  "NeZha",
			Age:   int32(3),
			Color: ColorYellow,
		}))
		Ω(obj).Should(Respect(&Person{ // Pointer type
			Name:  "NeZha",
			Age:   int32(3),
			Color: ColorYellow,
		}))
	})

	Context("Primitive", func() {
		It("Same with Equal matcher for primitive types", func() {
			Ω("a").Should(Respect("a"))
			Ω(true).Should(Respect(true))
			Ω(int32(3)).Should(Respect(int32(3)))
			Ω(0).Should(Respect(0))
			Ω(2.4).Should(Respect(2.4))
		})
	})

	Context("Struct", func() {
		It("Zero value will be ignored by default", func() {
			Ω(obj).Should(Respect(&Person{
				Name: "", // ignored
				//Age: int32(3), // ignored: Non-pointer field value will be zero if not provided
				Color: ColorYellow,
			}))
		})

		It("Should respect zero values if ZeroValueMatters option was set", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name: "NeZha",
				//Age: int32(3), // Non-pointer field value will be zero if not provided
				Color: ColorYellow,
			}, respect.ZeroValueMatters))
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3), // Required value provided
				Color: ColorYellow,
			}, respect.ZeroValueMatters))
		})
	})

	Context("Slice", func() {
		It("Slice of string items could have different order", func() {
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"right", "left"},
				},
			}))
		})

		It("Slice of struct items could have different order", func() {
			Ω(obj).Should(Respect(&Person{
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
			}))
		})

		It("Slice of struct should provide valid/non-zero string/*string field identifier", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Legs: []*Leg{
						{
							//Name: &LegRight,
						},
						{
							Name: &LegLeft,
						},
					},
				},
			}))
		})

		It("Slice items should have same orders if OrderMatters option set", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"right", "left"}, // Wrong order
				},
			}, respect.OrderMatters))
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"left", "right"}, // Correct order
				},
			}, respect.OrderMatters))
		})

		It("Slice in respectObj can provide less items but shouldn't provide more", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"left", "right", "mix"}, // More items
				},
			}))
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"left"}, // Less items
					Legs: []*Leg{
						{
							Name: &LegRight,
						},
					},
				},
			}))
		})

		It("Slice can't provide less items if has LengthMatters option set", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"left"}, // Less items
				},
			}, respect.LengthMatters))
		})
	})
})
