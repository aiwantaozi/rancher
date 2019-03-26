package client

const (
	InjectAbortType            = "injectAbort"
	InjectAbortFieldHttpStatus = "httpStatus"
	InjectAbortFieldPerecent   = "percent"
)

type InjectAbort struct {
	HttpStatus int64 `json:"httpStatus,omitempty" yaml:"httpStatus,omitempty"`
	Perecent   int64 `json:"percent,omitempty" yaml:"percent,omitempty"`
}
