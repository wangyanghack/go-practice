package main

import (
	"bytes"
	"fmt"
	"reflect"
)

type order struct {
	ordId      int
	customerId int
}

type employee struct {
	name    string
	id      int
	address string
	salary  int
	country string
}

func main() {
	o := order{
		ordId:      456,
		customerId: 56,
	}
	createQuery(o)

	e := employee{
		name:    "Naveen",
		id:      565,
		address: "Coimbatore",
		salary:  90000,
		country: "India",
	}
	createQuery(e)
	i := 90
	createQuery(i)
}

func createQuery(q interface{}) {
	value := reflect.ValueOf(q)
	if value.Kind() != reflect.Struct {
		fmt.Println("unsupported type")
		return
	}
	var sql bytes.Buffer
	sql.WriteString(fmt.Sprintf("insert into %s values(", reflect.TypeOf(q).Name()))
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Kind() == reflect.Int {
			sql.WriteString(fmt.Sprintf("%d", value.Field(i).Int()))
		} else if value.Field(i).Kind() == reflect.String {
			sql.WriteString(fmt.Sprintf("\"%s\"", value.Field(i).String()))
		}
		if i == value.NumField()-1 {
			sql.WriteString(")")
		} else {
			sql.WriteString(", ")
		}
	}
	fmt.Println(sql.String())
}
