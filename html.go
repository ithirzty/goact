package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

type element struct {
	Name     string
	Attrs    map[string]string
	Content  string
	Children []int
	Parent   int
	Indent   int
	Id       int
}

var eId int = 0

func parseHTML(content string) string {
	if len(currentWriterName) == 0 {
		fmt.Println("Missing http context when echoing html")
		os.Exit(1)
	}
	lines := strings.Split(content, "\n")
	elems := []element{}
	if len(lines) < 2 {
		fmt.Println("Need one element per LINE")
		os.Exit(1)
	}
	lines = lines[1:]
	baseIndent := countIndent(lines[0])

	for _, l := range lines {
		if len(strings.TrimSpace(l)) == 0 {
			continue
		}

		lastElement := element{Indent: baseIndent}
		if len(elems) > 0 {
			lastElement = elems[len(elems)-1]
		}
		elems = append(elems, parseElem(l))
		tmpElement := elems[len(elems)-1]

		//is a child
		if tmpElement.Indent > lastElement.Indent {
			lastElement.Children = append(lastElement.Children, tmpElement.Id)
			tmpElement.Parent = lastElement.Id
			elems[len(elems)-1] = tmpElement
			elems[len(elems)-2] = lastElement
			continue
		} else if tmpElement.Indent != baseIndent {
			parent := getParent(tmpElement, lastElement, elems)
			for i, e := range elems {
				if e.Id == parent {
					tmpElement.Parent = parent
					elems[len(elems)-1] = tmpElement
					e.Children = append(e.Children, tmpElement.Id)
					elems[i] = e
					break
				}
			}
		}

	}

	code := currentWriterName + ".Write([]byte(`" + marshalElems(elems, baseIndent) + "`))\n"
	return code
}

func countIndent(line string) int {
	for i, c := range line {
		if c != '	' {
			return i
		}
	}
	return 0
}

func parseElem(l string) element {
	elem := element{}
	eId++
	elem.Id = eId
	elem.Indent = countIndent(l)
	l = strings.TrimSpace(l)
	l += "\n"
	i := 0
	elem.Attrs = map[string]string{}
	//getting elem name
	for ; unicode.IsLetter(rune(l[i])) || unicode.IsNumber(rune(l[i])) || rune(l[i]) == '-' || rune(l[i]) == '_'; i++ {
		elem.Name += string(l[i])
	}

	for i < len(l) {
		if l[i] == '.' {
			//assing class
			i++
			for ; unicode.IsLetter(rune(l[i])) || unicode.IsNumber(rune(l[i])) || rune(l[i]) == '-' || rune(l[i]) == '_'; i++ {
				elem.Attrs["\"class\""] += string(l[i])
			}
			elem.Attrs["\"class\""] += " "
		} else if l[i] == '#' {
			//assign id
			i++
			for ; unicode.IsLetter(rune(l[i])) || unicode.IsNumber(rune(l[i])) || rune(l[i]) == '-' || rune(l[i]) == '_'; i++ {
				elem.Attrs["\"id\""] += string(l[i])
			}
		} else if l[i] == '{' {
			//assign multiple attributes
			i++

			//get json data
			isString := false
			isEscaped := false
			memory := []rune{}
			nbBraces := 1

			for _, c := range l[i:] {

				//escaping char
				if isEscaped {
					memory = append(memory, c)
					isEscaped = false
					continue
				}

				//need to escape following char
				if c == '\\' {
					memory = append(memory, c)
					isEscaped = true
					continue
				}

				//current token is a string
				if isString {
					if c == '"' {
						isString = false
						memory = append(memory, c)
						continue
					}
					memory = append(memory, c)
					continue
				}

				//current token is not yet a string
				if c == '"' {
					memory = append(memory, c)
					isString = true
					continue
				}

				//is end of json
				if c == '{' {
					nbBraces++
				} else if c == '}' {
					nbBraces--
				}
				if nbBraces == 0 {
					break
				}

				//add all remaining chars
				memory = append(memory, c)

			}

			//parse json to attricutes

			jsonAttrs := parseJson(memory)
			for k, v := range jsonAttrs {
				elem.Attrs[k] = v
			}

		} else if l[i] == '=' {
			i++
			break
		} else {
			i++
		}

	}
	//getting after the equal sign
	for i < len(l) {
		elem.Content += string(l[i])
		i++
	}

	elem.Attrs["\"class\""] = "\"" + elem.Attrs["\"class\""] + "\""
	elem.Attrs["\"id\""] = "\"" + elem.Attrs["\"id\""] + "\""

	elem.Content = strings.TrimSpace(elem.Content)

	return elem
}

func parseJson(content []rune) map[string]string {
	attrs := map[string]string{}

	isString := false
	isEscaped := false
	key := ""
	val := ""
	isKey := true

	for _, c := range content {

		//escaping char
		if isEscaped {
			if isKey {
				key += string(c)
			} else {
				val += string(c)
			}
			isEscaped = false
			continue
		}

		//need to escape following char
		if c == '\\' {
			if isKey {
				key += string(c)
			} else {
				val += string(c)
			}
			isEscaped = true
			continue
		}

		//current token is a string
		if isString {
			if c == '"' {
				isString = false
				if isKey {
					key += string(c)
				} else {
					val += string(c)
				}
				continue
			}
			if isKey {
				key += string(c)
			} else {
				val += string(c)
			}
			continue
		}

		//current token is not yet a string
		if c == '"' {
			if isKey {
				key += string(c)
			} else {
				val += string(c)
			}
			isString = true
			continue
		}

		//expecting value after key
		if c == ':' {
			if isKey {
				isKey = false
			} else {
				fmt.Println("Cannot add two values to a single key '" + key + "'")
				os.Exit(1)
			}
			continue
		}

		//getting to the next pair
		if c == ',' {
			if isKey {
				fmt.Println("Missing value for key '" + key + "'")
				os.Exit(1)
			} else {
				attrs[key] = val
				key = ""
				val = ""
				isKey = true
			}
			continue
		}

		//adding all other chars
		if isKey {
			key += string(c)
		} else {
			val += string(c)
		}

	}
	attrs[key] = val

	return attrs
}

func marshalElems(elems []element, baseIndent int) string {
	code := ""
	for _, e := range elems {
		if e.Indent != baseIndent {
			continue
		}

		attrs := ""
		for k, v := range e.Attrs {
			attrs += "`+" + k + "+`="
			attrs += "\"`+" + v + "+`\" "
		}

		code += "<" + e.Name + " " + attrs + ">"

		if len(e.Content) > 0 {
			code += "`+" + e.Content + "+`"
		}

		children := getAllChildren(elems, e)

		code += marshalElems(children, e.Indent+1)

		code += "</" + e.Name + ">"

	}
	return code
}

func getAllChildren(elems []element, e element) []element {
	els := []element{e}
	for _, c := range e.Children {
		for _, ce := range elems {
			if ce.Id == c {
				els = append(els, getAllChildren(elems, ce)...)
			}
		}
	}
	return els
}

func getParent(tmpElem, lastElem element, elems []element) int {
	for tmpElem.Indent != lastElem.Indent {
		lastElem = getElementById(lastElem.Parent, elems)
	}
	return lastElem.Parent
}

func getElementById(id int, elems []element) element {
	for _, e := range elems {
		if e.Id == id {
			return e
		}
	}
	return element{}
}
