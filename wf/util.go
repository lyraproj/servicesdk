package wf

import (
	"errors"
	"fmt"
)

func ToError(r interface{}) error {
	switch rx := r.(type) {
	case error:
		return rx
	case fmt.Stringer:
		return errors.New(rx.String())
	case string:
		return errors.New(rx)
	default:
		return fmt.Errorf("%+v", r)
	}
}
