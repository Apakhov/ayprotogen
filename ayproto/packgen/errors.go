package packgen

import "github.com/pkg/errors"

var (
	ErrNotStruct     = errors.New("not struct")
	ErrBadType       = errors.New("bad type")
	ErrNoStructName  = errors.New("no name")
	ErrPtrNotAllowed = errors.New("poiters ot allowed")

	ErrNotFunc      = errors.New("not func")
	ErrFuncSiganure = errors.New("function not of type func (context.Context, in) (out, err)")
)

type StructErr struct {
	Err  error
	Name string
}
