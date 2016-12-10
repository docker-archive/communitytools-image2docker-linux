package main

import (
	"errors"
	"fmt"
	"github.com/docker/v2c/workflow"
	"github.com/urfave/cli"
	"os"
)

const name string = `v2c`
const description string = `Lift and shift the contents of a virtual machine image
         into build materials for a Docker image.`
const version string = `v0.1.0`

var (
	errAtLeastOne  = errors.New(`expected at least one argument`)
	errExactlyOne  = errors.New(`expected exactly one argument`)
	errExactlyNone = errors.New(`no arguments are expected`)
)

func main() {
	newApp().Run(os.Args)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = name
	app.Usage = description
	app.Version = version
	app.Commands = []cli.Command{
		{
			Name:     `build`,
			Usage:    `transform a virtual disk into a container`,
			Category: `Transform`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  `tag, t`,
					Usage: "Tag the resulting image with `REPOSITORY[:TAG]`",
				},
				cli.StringSliceFlag{
					Name: `label, l`,
				},
			},
			Action: buildHandler,
		},
		{
			Name:     `image`,
			Usage:    `options for working with transformed images`,
			Category: `Transform`,
			Subcommands: []cli.Command{
				{
					Name:   `rm`,
					Usage:  `remove a transformed image`,
					Action: removeImageHandler,
				},
				{
					Name:   `list`,
					Usage:  `list the transformed images`,
					Action: listImageHandler,
				},
				{
					Name:  `export`,
					Usage: `export a transformed image`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  `output, o`,
							Usage: "Write to a `file` instad of STDOUT",
						},
					},
					Action: exportImageHandler,
				},
			},
		},
		{
			Name:     `detective`,
			Usage:    `options for working with detectives`,
			Category: `Component`,
			Subcommands: []cli.Command{
				{
					Name:   `install`,
					Usage:  `install a detective`,
					Action: installDetectiveHandler,
				},
				{
					Name:   `list`,
					Usage:  `list the installed detectives`,
					Action: listDetectiveHandler,
				},
			},
		},
		{
			Name:     `provisioner`,
			Usage:    `options for working with provisioners`,
			Category: `Component`,
			Subcommands: []cli.Command{
				{
					Name:   `install`,
					Usage:  `install a provisioner`,
					Action: installProvisionerHandler,
				},
				{
					Name:   `list`,
					Usage:  `list the installed provisioners`,
					Action: listProvisionerHandler,
				},
			},
		},
	}
	return app
}

func defaultHandler(c *cli.Context) error {
	err := workflow.Start()
	if err != nil {
		panic(err)
	}
	fmt.Println("BOOM")

	return nil
}

func buildHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Running image transformation.")
	return nil
}

func listDetectiveHandler(c *cli.Context) error {
	if c.NArg() > 0 {
		return errExactlyNone
	}

	type detective struct {
		Repository  string
		Tag         string
		Category    string
		Description string
	}
	renderTabbed(`detectiveList`, os.Stdout, struct {
		Detectives []detective
	}{
		Detectives: []detective{
			detective{
				Repository:  `repo`,
				Tag:         `tag`,
				Category:    `original`,
				Description: `created`,
			},
			detective{
				Repository:  `repo2`,
				Tag:         `taggert`,
				Category:    `crap`,
				Description: `Some time in june`,
			},
		},
	})
	return nil
}

func installDetectiveHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Installing a detective")
	return nil
}

func listProvisionerHandler(c *cli.Context) error {
	if c.NArg() > 0 {
		return errExactlyNone
	}

	type provisioner struct {
		Repository  string
		Tag         string
		Category    string
		Description string
	}
	renderTabbed(`provisionerList`, os.Stdout, struct {
		Provisioners []provisioner
	}{
		Provisioners: []provisioner{
			provisioner{
				Repository:  `repo`,
				Tag:         `tag`,
				Category:    `original`,
				Description: `created`,
			},
			provisioner{
				Repository:  `repo2`,
				Tag:         `taggert`,
				Category:    `crap`,
				Description: `Some time in june`,
			},
		},
	})
	return nil
}

func installProvisionerHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Installing provisioners")
	return nil
}

func removeImageHandler(c *cli.Context) error {
	if c.NArg() < 1 {
		return errAtLeastOne
	}
	fmt.Println("Removing an image")
	return nil
}

func listImageHandler(c *cli.Context) error {
	if c.NArg() > 0 {
		return errExactlyNone
	}

	type image struct {
		ID         string
		Repository string
		Tag        string
		Original   string
		Created    string
	}
	renderTabbed(`imageList`, os.Stdout, struct {
		Images []image
	}{
		Images: []image{
			image{
				ID:         `anid`,
				Repository: `repo`,
				Tag:        `tag`,
				Original:   `original`,
				Created:    `created`,
			},
			image{
				ID:         `dina`,
				Repository: `repo2`,
				Tag:        `taggert`,
				Original:   `crap`,
				Created:    `Some time in june`,
			},
		},
	})
	return nil
}

func exportImageHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Exporting an image")
	return nil
}
