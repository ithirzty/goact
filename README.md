# What is goact
Goact is a better way of pushing html to the client.

## How to use it

Execute the binary inside your project folder.
Linux
```bash
./goact
```
Windows
```cmd
.\goact.exe
```
### The go compiler needs to be installed.


Here is a test program:

```golang
package main

import (
	"net/http"
	"strconv"
)


func main() {
	http.HandleFunc("/", server)
	http.ListenAndServe(":80", nil)
}

var i int = 0

func server(w http.ResponseWriter, r *http.Request) {

i++

echo(
    html
        head
	
    	body
            div#content
                h1 = "It works!"
            footer
                p{"style":"color:red"} = "Visitors: "+strconv.Itoa(i)
)
}
```

### Notice something strange?
> Instead of using ``w.Write(`<html>...</html>`)`` we are using ``echo()``.

## Syntax
Each line is an element. The hiearchy is done by indenting lines, childs of an element will be indented with one more tabulation.
__The indentation only works with tabulations.__
You can add one information per element and a content.

### Attributes

```
element#id
```
Will set an id to the element.

```
element.class
```
Will set a class to the element.

##### For additional attributes:
```json
element{"class": "class1 class2", "alt": "I'm an element."}
``` 
Where not put in `"`, key and values can be go variables.

### Content
```golang
element = "some text"+someVariable
```
Will set the content of an element to the quoted text or the value of the passed variable.
> Can I use functions?
__Yes.__
##### As such:
```golang
i++
echo (
    p = "Number of visitors: "
        span{"color":"red"} = fmt.Sprint(i)
)
```

## What's new?

<blockquote>Support for multiple attribute per element ``div.col.row#test``</blockquote>