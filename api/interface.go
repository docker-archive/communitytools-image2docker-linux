package api

type Detective struct {
	ImageID     string
	Repository  string
	Tag         string
	Category    string
	Description string
	Related     string
}

type Provisioner struct {
	ImageID     string
	Repository  string
	Tag         string
	Category    string
	Description string
}

type Packager struct {
	ImageID     string
	Repository  string
	Tag         string
	Category    string
	Description string
}

type Product struct {
	ImageID    string
	Repository string
	Tag        string
	Original   string
	Created    string
}
