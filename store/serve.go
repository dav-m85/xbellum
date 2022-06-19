package store

import (
	"html/template"
	"net/http"
)

var tplStr string = `
<html>
<body>
{{range .}}
<h2>{{.Version}} (from {{.ParentVersion}})</h2>
<p><b>Adds {{len .Adds}}</b></p>
<p><b>Removes {{len .Removes}}</b></p>
		{{range .Adds}}
		<p style="color: green">+ {{.Href}}</p>
		{{end}}
		{{range .Removes}}
		<p style="color: red">- {{.Href}}</p>
	{{end}}
{{end}}
</body>
</html>
`

func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tpl, _ := template.New("main").Parse(tplStr)

	diffs, _ := s.DiffAll()

	err := tpl.Execute(w, diffs)
	if err != nil {
		panic(err)
	}
}
