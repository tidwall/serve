# serve

Just a simple http server for Go

## Example

```go
package main

import "github.com/tidwall/serve"

func main() {
	serve.Serve(serve.Options{
		Handler: http.HandlerFunc(myHandler),
		Domain:  "example.com",
	})
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>HIYA!</h1>")
}
```

