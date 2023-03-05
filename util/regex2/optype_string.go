// Code generated by "stringer -type=opType"; DO NOT EDIT.

package regex2

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[opChar-1]
	_ = x[opCharIgnoreCase-2]
	_ = x[opJump-3]
	_ = x[opSplitFirst-4]
	_ = x[opSplitLast-5]
	_ = x[opAny-6]
	_ = x[opAnyNotNL-7]
	_ = x[opHalfSet-8]
	_ = x[opFullSet-9]
	_ = x[opListSet-10]
	_ = x[opWordStart-11]
	_ = x[opWordEnd-12]
	_ = x[opLineStart-13]
	_ = x[opLineEnd-14]
	_ = x[opStrStart-15]
	_ = x[opStrEnd-16]
	_ = x[opSave-17]
	_ = x[opDoneSave1-18]
	_ = x[opOnePass-19]
}

const _opType_name = "opCharopCharIgnoreCaseopJumpopSplitFirstopSplitLastopAnyopAnyNotNLopHalfSetopFullSetopListSetopWordStartopWordEndopLineStartopLineEndopStrStartopStrEndopSaveopDoneSave1opOnePass"

var _opType_index = [...]uint8{0, 6, 22, 28, 40, 51, 56, 66, 75, 84, 93, 104, 113, 124, 133, 143, 151, 157, 168, 177}

func (i opType) String() string {
	i -= 1
	if i >= opType(len(_opType_index)-1) {
		return "opType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _opType_name[_opType_index[i]:_opType_index[i+1]]
}
