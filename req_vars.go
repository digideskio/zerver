package zerver

import "net/url"

type ReqVars struct {
	urlVars   map[string]int
	urlVals   []string
	queryVars url.Values
	formVars  url.Values
}

// URLVar return values of variable
func (v *ReqVars) URLVar(name string) string {
	if v.urlVars == nil {
		return ""
	}
	i, has := v.urlVars[name]
	if !has {
		return ""
	}
	return v.urlVals[i]
}

func (v *ReqVars) QueryVar(name string) string {
	if v.queryVars == nil {
		return ""
	}
	return v.queryVars.Get(name)
}

func (v *ReqVars) QueryVarMul(name string) []string {
	if v.queryVars == nil {
		return nil
	}
	return v.queryVars[name]
}

func (v *ReqVars) FormVar(name string) string {
	if v.formVars == nil {
		return ""
	}
	return v.formVars.Get(name)
}

func (v *ReqVars) FormVarMul(name string) []string {
	if v.formVars == nil {
		return nil
	}
	return v.formVars[name]
}
