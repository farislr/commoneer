package pkgservice

type Code int

const (
	ErrInternal Code = iota
)

var codeToDetailMap = map[Code]string{
	ErrInternal: "internal_error",
}

func (c Code) String() string {
	return codeToDetailMap[c]
}
