package test_test

import (
	"fmt"
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
	Name *string
}

type Body struct {
	Head *Head
	Arms []string
	Legs []*Leg
}

type Person struct {
	Name        string
	Age         int32
	description string `json:"description,omitempty"`
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
)

var _ = BeforeSuite(func() {
	obj = &Person{
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
})

var _ = Describe("Test", func() {

	BeforeEach(func() {
		fmt.Println(CurrentGinkgoTestDescription().FullTestText)
	})

	It("Pointer", func() {
		Ω(obj).ShouldNot(Respect(Person{
			Name:  "NeZha",
			Age:   int32(3),
			Color: ColorYellow,
		}))
		Ω(obj).Should(Respect(&Person{
			Name:  "NeZha",
			Age:   int32(3),
			Color: ColorYellow,
		}))
	})

	Context("Premitive", func() {
		It("", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name:  "NeZhaFake",
				Age:   int32(2),
				Color: ColorBlack,
			}))
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
			}))
		})
	})

	Context("Struct", func() {
		It("Required value must be provided to avoid unexpected result", func() {
			Ω(obj).ShouldNot(Respect(&Person{
				Name: "NeZha",
				//Age: int32(3), // Required field value will be zero if not provided
				Color: ColorYellow,
			}))
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3), // Required value provided
				Color: ColorYellow,
			}))
		})
	})

	Context("Slice", func() {
		It("Slice items have different order", func() {
			Ω(obj).Should(Respect(&Person{
				Name:  "NeZha",
				Age:   int32(3),
				Color: ColorYellow,
				Body: Body{
					Arms: []string{"right", "left"},
				},
			}))
		})

		It("Struct slice have different order", func() {
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

		It("Slice items should have same orders if  OrderMatters option set", func() {
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

		It("Slice can provide less items but shouldn't provide more", func() {
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
