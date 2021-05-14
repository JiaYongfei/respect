package respect

//Options is the type for options passed to respect matcher.
type Options int

const (
	//IgnoreExtras tells the matcher to ignore extra elements or fields, rather than triggering a failure.
	IgnoreExtras Options = 1 << iota
	//IgnoreMissing tells the matcher to ignore missing elements or fields, rather than triggering a failure.
	IgnoreMissing
	//AllowDuplicates tells the matcher to permit multiple members of the slice to produce the same ID when
	//considered by the indentifier function. All members that map to a given key must still match successfully
	//with the matcher that is provided for that key.
	AllowDuplicates
	//OrderMatters will take the item order into consideration when compare slices or arrays
	OrderMatters
	//LengthMatters will consider the length of array and slice when compare
	LengthMatters
)
