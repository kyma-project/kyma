// Code generated by "stringer -type=ConditionStatus -trimprefix=ConditionStatus"; DO NOT EDIT.

package controllers

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ConditionStatusSucceeded-1]
	_ = x[ConditionStatusFailed-2]
	_ = x[ConditionStatusUnknown-3]
}

const _ConditionStatus_name = "SucceededFailedUnknown"

var _ConditionStatus_index = [...]uint8{0, 9, 15, 22}

func (i ConditionStatus) String() string {
	i -= 1
	if i < 0 || i >= ConditionStatus(len(_ConditionStatus_index)-1) {
		return "ConditionStatus(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ConditionStatus_name[_ConditionStatus_index[i]:_ConditionStatus_index[i+1]]
}
