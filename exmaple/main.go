package main

import (
	"encoding/json"
	"fmt"

	"github.com/blues/jsonata-go"
	"github.com/blues/jsonata-go/jparse"
)

// const rule = `(
//
// $gender:= original_person.gender;
//
//	{
//	  "var_name": name,
//	  "var_null": null,
//	  "var_age": age+1 + gender,
//	  "var_feat":{
//	    "var_sex": sex
//	  }
//	})`
// const rule = `(
// {
//   "var_name": name,
//   "var_name2": original_person.addr.line
// })`

const rule = `
$formatInteger(23, "#,0")

`

const inputStr = `
{

}


`

func main() {
	var input interface{}
	err := json.Unmarshal([]byte(inputStr), &input)
	if err != nil {
		panic(err)
	}

	// expression := jsonata.MustCompile("$sum(example.value)")

	node, err := jparse.Parse(rule)
	if err != nil {
		panic(err)
	}
	out, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	expression := jsonata.MustCompile(rule)
	result, err := expression.Eval(input)
	if err != nil {
		fmt.Println("eval err", err)
	}
	out, err = json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}
