package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type token struct {
	Content string
	Type    int
}

var (
	CODE   = 0
	STRING = 1
	HTML   = 2
)

var currentWriterName string = ""

func main() {

	deleteFlag := flag.Bool("delete", false, "Do not delete tmp files.")

	flag.Parse()

	files, _ := ioutil.ReadDir("./")
	tmpGOXfiles := []string{}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".go") {
			continue
		}
		content, _ := ioutil.ReadFile(f.Name())
		newFile := handleFile(string(content))

		tmpGOXfiles = append(tmpGOXfiles, f.Name())
		if runtime.GOOS != "windows" {
			os.Create("TMPGOX_" + f.Name())
		}
		ioutil.WriteFile("TMPGOX_"+f.Name(), []byte(newFile), os.FileMode(os.O_CREATE))
		os.Rename(f.Name(), f.Name()+"x")
	}

	cmd := exec.Command("go", "build", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	for _, n := range tmpGOXfiles {
		os.Rename(n+"x", n)
		if *deleteFlag != true {
			os.Remove("TMPGOX_" + n)
		}
	}

}

func handleFile(content string) string {
	isString := false
	isCode := 0
	isEscaped := false
	memory := []rune{}
	tokens := []token{}

	for _, c := range content {

		//escaping char
		if isEscaped {
			memory = append(memory, c)
			isEscaped = false
			continue
		}

		//need to escape following char
		if c == '\\' {
			if isString {
				memory = append(memory, c)
			}
			isEscaped = true
			continue
		}

		//current token is html
		if isCode > 0 {
			if c == '"' {
				if isString {
					isString = false
				} else {
					isString = true
				}
			}
			if c == '(' && !isString {
				isCode++
			} else if c == ')' && !isString {
				isCode--
			}
			if isCode == 0 {
				tokens = append(tokens, token{Type: HTML, Content: string(memory)})
				memory = []rune{}
				continue
			}
			memory = append(memory, c)
			continue
		}

		//current token is not yet html
		if c == '(' && strings.HasSuffix(strings.TrimSpace(string(memory)), "echo") {
			isCode = 1
			code := " " + strings.TrimSpace(string(memory))
			code = code[:len(code)-4]
			tokens = append(tokens, token{Type: CODE, Content: code})
			memory = []rune{}
			continue
		}

		//current token is a string
		if isString {
			if c == '"' {
				isString = false
				memory = append(memory, c)
				tokens = append(tokens, token{Content: string(memory), Type: STRING})
				memory = []rune{}
				continue
			}
			memory = append(memory, c)
			continue
		}

		//current token is not yet a string
		if c == '"' {
			tokens = append(tokens, token{Content: string(memory), Type: CODE})
			memory = []rune{c}
			isString = true
			continue
		}

		//is actually code
		memory = append(memory, c)

	}

	tokens = append(tokens, token{Type: CODE, Content: string(memory)})

	//fmt.Println(tokens)

	return parse(tokens)

}

func parse(tokens []token) string {
	code := ""
	for _, t := range tokens {
		switch t.Type {
		case CODE:
			code += parseCode(t.Content)
		case STRING:
			code += t.Content
		case HTML:
			code += parseHTML(t.Content)
		}

	}
	return code
}

func parseCode(content string) string {
	lines := strings.Split(content, "\n")
	for _, l := range lines {

		//not declaring a func
		if !strings.HasPrefix(strings.TrimSpace(l), "func ") && !strings.Contains(l, "http.ResponseWriter") {
			continue
		}

		vnRegex := regexp.MustCompile(`(\w*) (http\.ResponseWriter)`)
		vns := vnRegex.FindAllSubmatch([]byte(l), 1)
		if len(vns) > 0 {
			currentWriterName = string(vns[0][1])
		} else {
			currentWriterName = ""
		}
	}

	return strings.Join(lines, "\n")
}
