package a

func basicSliceComparison() {
	var s []int
	if s == nil { // want "slice compared to nil"
		// should flag this
	}
	if s != nil { // want "slice compared to nil"
		// should flag this
	}
}

func ignoredComparison() {
	var s []int
	//nillinter:ignore
	if s == nil {
		// should not flag this
	}
	//nillinter:ignore
	if s != nil {
		// should not flag this
	}
}

func nonSliceComparison() {
	var i *int
	if i == nil {
		// should not flag pointer comparison
	}
	var m map[string]int
	if m == nil {
		// should not flag map comparison
	}
	var ch chan int
	if ch == nil {
		// should not flag channel comparison
	}
}

func sliceInStruct() {
	type S struct {
		slice []int
	}
	var s S
	if s.slice == nil { // want "slice compared to nil"
		// should flag this
	}
}

func sliceInArray() {
	var arr [][]int
	if arr[0] == nil { // want "slice compared to nil"
		// should flag this
	}
}

func sliceFromFunction() {
	f := func() []int { return nil }
	if f() == nil { // want "slice compared to nil"
		// should flag this
	}
}

func nestedComparisons() {
	var s []int
	var t []string
	if s == nil && t != nil { // want "slice compared to nil" "slice compared to nil"
		// should flag both
	}
}

func nilOnLeft() {
	var s []int
	if nil == s { // want "slice compared to nil"
		// should flag this
	}
	if nil != s { // want "slice compared to nil"
		// should flag this
	}
}
