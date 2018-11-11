package wfapi

type Workflow interface {
	Activity

	Activities() []Activity
}
