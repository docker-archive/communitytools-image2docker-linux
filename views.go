package main

import (
	"github.com/docker/docker/api/types"
	"io"
	"text/tabwriter"
	"text/template"
)

func render(name string, out io.Writer, data interface{}) error {
	t := template.Must(template.New(name).Funcs(funcMap).Parse(views[name]))
	return t.Execute(out, data)
}

func renderTabbed(name string, out io.Writer, data interface{}) error {
	t := template.Must(template.New(name).Funcs(funcMap).Parse(views[name]))
	w := tabwriter.NewWriter(out, 16, 16, 16, ' ', 0)
	err := t.Execute(w, data)
	if err == nil {
		w.Flush()
	}
	return err
}

func renderRemoved(out io.Writer, gone []types.ImageDelete, ie error) error {
	t := template.Must(template.New(`removedImage`).Funcs(funcMap).Parse(views[`removedImage`]))
	w := tabwriter.NewWriter(out, 16, 16, 16, ' ', 0)
	err := t.Execute(w, struct {
		Gone []types.ImageDelete
	}{gone})
	if err == nil {
		w.Flush()
	} else {
		return err
	}
	return ie
}

var funcMap = template.FuncMap{
	"tail8":    func(c string) string { return c[len(c)-8:] },
	"tail12":   func(c string) string { return c[len(c)-12:] },
	"head8":    func(c string) string { return c[:8] },
	"head12":   func(c string) string { return c[:12] },
	"stripSha": func(c string) string { return c[7:] },
	"orNone": func(c string) string {
		if len(c) == 0 {
			return `-`
		}
		return c
	},
}

var views = map[string]string{
	`imageList`: `ID	REPOSITORY	TAG	ORIGINAL	CREATED{{ range .Products }}
{{.ImageID | stripSha | head12}}	{{.Repository}}	{{.Tag}}	{{.Original}}	{{.Created}}{{ end }}
`,
	`detectiveList`: `REPOSITORY	TAG	CATEGORY	DESCRIPTION{{ range .Detectives }}
{{.Repository}}	{{.Tag}}	{{.Category}}	{{.Description}}{{ end }}
`,
	`provisionerList`: `REPOSITORY	TAG	CATEGORY	DESCRIPTION{{ range .Provisioners }}
{{.Repository}}	{{.Tag}}	{{.Category}}	{{.Description}}{{ end }}
`,
	`removedImage`: `UNTAGGED	DELETED{{ range .Gone }}
{{.Untagged | orNone}}	{{.Deleted | orNone }}{{ end }}
`,
}
