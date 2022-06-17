package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/dav-m85/xbellum/vfs"
	"github.com/dav-m85/xbellum/xbel"
	"golang.org/x/net/webdav"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var dead bool
	flag.BoolVar(&dead, "c", false, "check for dead")
	flag.Parse()

	args := flag.Args()
	// if len(args) != 2 {
	// 	fmt.Println("Usage: go run main.go [-c] <input> <output>")
	// 	flag.PrintDefaults()
	// 	os.Exit(1)
	// }

	switch args[0] {
	case "server":
		wh := webdav.Handler{
			FileSystem: vfs.VFS{},
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, e error) {
				log.Printf("%s %s ERR:%s", r.Method, r.URL, e)
			},
		}
		listener, err := net.Listen("tcp", "127.0.0.1:8082")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Serving on %s", listener.Addr())
		if err := http.Serve(listener, &wh); err != nil {
			log.Print("shutting server", err)
		}
	}

	// nx := walk(xbel, func(b Bookmark) bool {
	// 	fmt.Println(b.Href)
	// 	return true
	// })

	// Remove duplicates
}

func inplace(input, output string) {
	buf, err := ioutil.ReadFile(input)
	check(err)

	x, err := xbel.Parse(buf)

	hrefs := make(map[string]struct{})
	nx := xbel.Walk(x, func(b *xbel.Bookmark) bool {
		if _, exists := hrefs[b.Href]; exists {
			fmt.Println("Duplicate " + b.Href)
			return false
		}

		hrefs[b.Href] = struct{}{}

		// Lets findout if still ok
		if false && strings.HasPrefix(b.Href, "https://example.com") {
			fmt.Println("Checking " + b.Href)
			resp, err := http.Get(b.Href)
			// handle the error if there is one
			if err != nil {
				panic(err)
			}
			// do this now so it won't be forgotten
			defer resp.Body.Close()
			// reads html as a slice of bytes
			html, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			// show the HTML code as a string %s
			if strings.Contains(string(html), "Post Not Found") {
				fmt.Println("DEAD")
				b.Title = "[DEAD] " + b.Title
			}
		}

		return true
	})

	outFile, err := os.Create(output)
	check(err)
	defer outFile.Close()

	xbel.Write(outFile, nx)
}
