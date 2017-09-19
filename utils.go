package main

func Yaml2JSON(i interface{}) interface{} {
	// https://stackoverflow.com/a/40737676/141895
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = Yaml2JSON(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = Yaml2JSON(v)
		}
	}
	return i
}
