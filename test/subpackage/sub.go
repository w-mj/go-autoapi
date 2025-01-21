package subpackage

import (
	"autoapi/test/model"
	m2 "autoapi/test/model"
	m3 "autoapi/test/model/subm"
	"fmt"
)

// @AutoAPI
func Add(data model.Model, d3 m3.Model) int {
	d2 := m2.Model{}
	return data.A + d2.B
}

func Func2() {
	fmt.Println("func2")
}
