package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
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

var (
	//flags of current running mode
	deleteFlag *bool
	testFlag   *bool

	//keep state of file for recompiling in test mode
	fileModifs = map[string]string{}

	//verbose progress
	progressLongest = 0
	lastFileName    = ""

	//binary handler for test mode
	binaryCmd *exec.Cmd
)

func main() {

	deleteFlag = flag.Bool("delete", false, "Do not delete tmp files.")
	testFlag = flag.Bool("test", false, "Run binary in test mode.")
	flag.Parse()

	if *testFlag {
		handleDir()
		dir, _ := os.Getwd()
		parent := filepath.Base(dir)
		if runtime.GOOS == "windows" {
			binaryCmd = exec.Command(`.\` + parent + ".exe")
		} else {
			binaryCmd = exec.Command(`./` + parent)
		}
		go func() {
			err := binaryCmd.Start()
			if err != nil {
				panic(err)
			}
		}()
		for {

			time.Sleep(time.Second * 1)
			changed := false
			newFileModifs := map[string]time.Time{}

			files, _ := ioutil.ReadDir("./")
			for _, f := range files {
				if !strings.HasSuffix(f.Name(), ".go") {
					continue
				}
				if fileModifs[f.Name()] != f.ModTime().String() {
					fmt.Println("Edited:", f.Name())
					changed = true
				}
				newFileModifs[f.Name()] = f.ModTime()
			}
			if changed {
				handleDir()
				binaryCmd.Process.Kill()
				dir, _ := os.Getwd()
				parent := filepath.Base(dir)
				if runtime.GOOS == "windows" {
					binaryCmd = exec.Command(`.\` + parent + ".exe")
				} else {
					binaryCmd = exec.Command(`./` + parent)
				}
				go func() {
					err := binaryCmd.Start()
					if err != nil {
						panic(err)
					}
				}()
			}

		}
	} else {
		handleDir()
	}

}

func handleDir() {
	files, _ := ioutil.ReadDir("./")
	tmpGOXfiles := []os.FileInfo{}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".go") {
			continue
		}

		content, _ := ioutil.ReadFile(f.Name())
		lastFileName = f.Name()
		newFile := handleFile(string(content))

		tmpGOXfiles = append(tmpGOXfiles, f)
		if runtime.GOOS != "windows" {
			os.Create("TMPGOX_" + f.Name())
		}
		ioutil.WriteFile("TMPGOX_"+f.Name(), []byte(newFile), os.FileMode(os.O_CREATE))
		os.Rename(f.Name(), f.Name()+"x")
		fmt.Print("\n")
	}

	cmd := exec.Command("go", "build", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()

	for _, f := range tmpGOXfiles {
		os.Rename(f.Name()+"x", f.Name())
		if *deleteFlag == false {
			os.Remove("TMPGOX_" + f.Name())
		}
		if *testFlag {
			fileModifs[f.Name()] = f.ModTime().String()
		}
	}
	fmt.Println("Done")
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
	for i, t := range tokens {
		progress(lastFileName, 10*i+1/len(tokens))
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

func progress(name string, step int) {
	p := name + " ["
	for i := 0; i <= 10; i++ {
		if i > step {
			p += "--"
		} else {
			p += "**"
		}
	}
	p += "]" + " "
	for i := 0; i <= progressLongest; i++ {
		if i > len(p) {
			p += " "
		}
	}
	progressLongest = len(p)
	fmt.Print("\r" + p)
}
