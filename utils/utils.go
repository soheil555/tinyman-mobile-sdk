package utils

import (
	b64 "encoding/base64"
	"sort"
	"strings"
)

type variable struct {
	Name string
	Type string
	Index int
	Length int
}

//TODO: definition type?
//TODO: definition['variables'] type?

/*   
	Return a byte array to be used in LogicSig.
*/
func GetProgram(definition map[string]interface{}, variables map[string]interface{}){

	template := definition["bytecode"].(string)
	templateBytes, _ := b64.StdEncoding.DecodeString(template)

	offset := 0

	var dVariables []variable = definition["variables"].([]variable)

	sort.SliceStable(variables, func(i,j int) bool {
		return dVariables[i].Index < dVariables[j].Index
	})


	for _,v := range dVariables {

		s := strings.Split(v.Name,"TMPL_")
		name := strings.ToLower(s[len(s)-1])
		value := variables[name]
		start := v.Index - offset
		end := start + v.Length


	}

}
