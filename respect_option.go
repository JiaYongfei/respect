package respect

//Options is the type for options passed to respect function/matcher.
type Options int

const (
	//OrderMatters will consider the items order when comparing array/slice, rather than triggering a failure.
	OrderMatters Options = 1 << iota
	//LengthMatters will consider the length of array/slice when comparing, rather than triggering a failure.
	LengthMatters
)
