package wfapi

type Action interface {
	Activity

	Function() interface{}
}
