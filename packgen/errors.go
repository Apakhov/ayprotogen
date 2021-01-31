package packgen

import (
	"errors"
	"fmt"
)

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

func errorWrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", msg, err.Error())
}
