package zerver

type patternKeeper interface {
	Pattern() string
}

type patternString string

func (ps patternString) Pattern() string {
	return string(ps)
}
