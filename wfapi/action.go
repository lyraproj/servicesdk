package wfapi

type Action interface {
	Activity

	Interface() interface{}
}
