package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/dav-m85/xbellum/store"
	"github.com/dav-m85/xbellum/vfs"
	"github.com/dav-m85/xbellum/xbel"
	"golang.org/x/crypto/bcrypt"
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

	st := store.NewStore()

	switch args[0] {
	case "server":
		// TODO generate password ans store it in file on first run
		saved := "secret"

		wh := webdav.Handler{
			FileSystem: vfs.NewVFS(st),
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, e error) {
				log.Printf("%s %s ERR:%s", r.Method, r.URL, e)
			},
		}
		listener, err := net.Listen("tcp", "127.0.0.1:8082")
		if err != nil {
			log.Fatal(err)
		}

		serve := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

			// Gets the correct user for this request.
			username, password, ok := r.BasicAuth()

			if !ok {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			if !checkPassword(saved, password) {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}

			_ = username
			// fmt.Println("Hello ", username)

			if r.URL.Path == "/info" {
				st.ServeHTTP(w, r)
			} else {
				wh.ServeHTTP(w, r)
			}
		}

		log.Printf("Serving on %s", listener.Addr())
		if err := http.Serve(listener, Server(serve)); err != nil {
			log.Print("shutting server", err)
		}

	case "dedup":
		buf, _ := st.Get()
		x, err := xbel.Parse(buf)
		if err != nil {
			log.Fatal(err)
		}

		hrefs := make(map[string]struct{})
		nx := xbel.Walk(x, func(b *xbel.Bookmark) bool {
			if _, exists := hrefs[b.Href]; exists {
				log.Println("Duplicate " + b.Href)
				return false
			}

			hrefs[b.Href] = struct{}{}

			return true
		})

		b := bytes.NewBuffer([]byte{})
		xbel.Write(b, nx)

		err = st.Set(b.Bytes())
		if err != nil {
			log.Fatal(err)
		}

	case "check":
		buf, _ := st.Get()
		x, err := xbel.Parse(buf)
		if err != nil {
			log.Fatal(err)
		}

		xbel.Walk(x, func(b *xbel.Bookmark) bool {
			// Lets findout if still ok
			// if false && strings.HasPrefix(b.Href, "https://example.com") {
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
			return true
		})
	}
}

type Server func(w http.ResponseWriter, r *http.Request)

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s(w, r)
}

func checkPassword(saved, input string) bool {
	if strings.HasPrefix(saved, "{bcrypt}") {
		savedPassword := strings.TrimPrefix(saved, "{bcrypt}")
		return bcrypt.CompareHashAndPassword([]byte(savedPassword), []byte(input)) == nil
	}

	return saved == input
}
