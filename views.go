package main

import (
	"io"
	"text/tabwriter"
	"text/template"
)

func render(name string, out io.Writer, data interface{}) error {
	t := template.Must(template.New(name).Parse(views[name]))
	return t.Execute(out, data)
}

func renderTabbed(name string, out io.Writer, data interface{}) error {
	t := template.Must(template.New(name).Parse(views[name]))
	w := tabwriter.NewWriter(out, 16, 16, 16, ' ', 0)
	err := t.Execute(w, data)
	if err == nil {
		w.Flush()
	}
	return err
}

var views = map[string]string{
	`imageList`: `ID	REPOSITORY	TAG	ORIGINAL	CREATED{{ range .Products }}
{{.ImageID}}	{{.Repository}}	{{.Tag}}	{{.Original}}	{{.Created}}{{ end }}
`,
	`detectiveList`: `REPOSITORY	TAG	CATEGORY	DESCRIPTION{{ range .Detectives }}
{{.Repository}}	{{.Tag}}	{{.Category}}	{{.Description}}{{ end }}
`,
	`provisionerList`: `REPOSITORY	TAG	CATEGORY	DESCRIPTION{{ range .Provisioners }}
{{.Repository}}	{{.Tag}}	{{.Category}}	{{.Description}}{{ end }}
`,
}
