package zerver

type errResp struct {
	Error interface{} `json:"error"`
}

var NewError = func(s interface{}) interface{} {
	return errResp{s}
}
