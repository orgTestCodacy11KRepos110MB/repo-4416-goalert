// Code generated by "stringer -type MessageType"; DO NOT EDIT.

package notification

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MessageTypeUnknown-0]
	_ = x[MessageTypeAlert-1]
	_ = x[MessageTypeAlertStatus-2]
	_ = x[MessageTypeTest-3]
	_ = x[MessageTypeVerification-4]
	_ = x[MessageTypeAlertBundle-5]
	_ = x[MessageTypeAlertStatusBundle-6]
}

const _MessageType_name = "MessageTypeUnknownMessageTypeAlertMessageTypeAlertStatusMessageTypeTestMessageTypeVerificationMessageTypeAlertBundleMessageTypeAlertStatusBundle"

var _MessageType_index = [...]uint8{0, 18, 34, 56, 71, 94, 116, 144}

func (i MessageType) String() string {
	if i < 0 || i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
