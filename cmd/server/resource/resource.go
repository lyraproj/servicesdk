package resource

import (
	"fmt"
)

type Foo struct {
}

func (*Foo) Hello(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}

type Bar struct {
}

func (*Bar) Hello(name string) (string, string) {
	return "Hello", name
}

type MyRes struct {
	Name  string
	Phone string
}

type CrdResource struct {
	Name string
	Age  int32
}
type CrdHandler struct {
}

// The Type of an exported API is reflected which means that the parameter types and return types
// are reflected too. They should not be of type interface when the type is known (which it should
// be)

func (c *CrdHandler) Create(desiredState *CrdResource) string {
	return "anExternalID"
}

func (c *CrdHandler) Read(externalID string) *CrdResource {
	return &CrdResource{
		Name: "readie",
		Age:  12,
	}
}

func (c *CrdHandler) Update(externalID string, desiredState *CrdResource) *CrdResource {
	return desiredState
}

func (c *CrdHandler) Delete(externalID string) error {
	return nil
}

func (c *CrdHandler) Delete2(externalID string) (string, error) {
	return "ok", nil
}

func (c *CrdHandler) Delete3(externalID string) (string, error) {
	return "", fmt.Errorf("not ok")
}
