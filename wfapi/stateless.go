package wfapi

type Stateless interface {
	Activity

	Interface() interface{}
}
