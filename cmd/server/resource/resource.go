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

func (c *CrdHandler) Create(desiredState interface{}) string {
	if _, ok := desiredState.(CrdResource); !ok {
		panic(fmt.Sprintf("desiredState was not an instance of CrdResource, it was %T", desiredState))
	}
	return "anExternalID"
}

func (c *CrdHandler) Read(externalID string) interface{} {
	return &CrdResource{
		Name: "readie",
		Age:  12,
	}
}

func (c *CrdHandler) Update(externalID string, desiredState interface{}) interface{} {
	return desiredState.(CrdHandler)
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
