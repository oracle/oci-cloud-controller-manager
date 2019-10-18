package apierrors

import "fmt"

func (e ErrorOCI) Error() string {
	output := e.Message

	if e.InnerError != nil {
		output += ": "
		if len(e.InnerError.Status) > 0 {
			output += fmt.Sprintf("Status: %s;", e.InnerError.Status)
		}
		if len(e.InnerError.Code) > 0 {
			output += fmt.Sprintf("Code: %s;", e.InnerError.Code)
		}
		if len(e.InnerError.RequestID) > 0 {
			output += fmt.Sprintf("OPC Request ID: %s; ", e.InnerError.RequestID)
		}
		if len(e.InnerError.Message) > 0 {
			output += fmt.Sprintf("Message: %s; ", e.InnerError.Message)
		}
		return output
	}

	if len(e.Status) > 0 || len(e.Code) > 0 || len(e.RequestID) > 0 {
		output += ": "
		if len(e.Status) > 0 {
			output += fmt.Sprintf("Status: %s;", e.Status)
		}
		if len(e.Code) > 0 {
			output += fmt.Sprintf("Code: %s;", e.Code)
		}
		if len(e.RequestID) > 0 {
			output += fmt.Sprintf("OPC Request ID: %s; ", e.RequestID)
		}
	}
	return output
}

func (e *ErrorOCI) ToV3() *ErrorV3 {
	r := &ErrorV3{}
	if e.InnerError != nil {
		r.Message = e.InnerError.Message
		r.Status = e.InnerError.Status
		r.Code = e.InnerError.Code
		r.OPCRequestID = e.InnerError.RequestID
		return r
	}

	r.Message = e.Message
	r.Status = e.Status
	r.Code = e.Code
	r.OPCRequestID = e.RequestID
	return r
}
