package wfapi

type Stateless interface {
	Activity

	Function() interface{}
}
