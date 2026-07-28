package model

import "encoding/json"

// StringOrBool handles a JSON field that the UFiber API exposes either as a
// string or as a boolean, depending on the platform.
//
// GPON OLTs return a string for SfpModule.TxFault, while XGS OLTs return a
// boolean. Since Go is strictly typed, this single mismatch makes the whole
// response fail to unmarshal: without this type, no metrics at all are
// collected from XGS devices.
//
// Boolean values are normalised to "true"/"false" so that any code already
// consuming this field as a string keeps working.
type StringOrBool string

func (s *StringOrBool) UnmarshalJSON(data []byte) error {
	// A JSON value not wrapped in quotes is not a string: try the boolean
	// form before falling back to the nominal case.
	if len(data) > 0 && data[0] != '"' {
		var b bool
		if err := json.Unmarshal(data, &b); err == nil {
			if b {
				*s = "true"
			} else {
				*s = "false"
			}
			return nil
		}
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = StringOrBool(str)
	return nil
}
