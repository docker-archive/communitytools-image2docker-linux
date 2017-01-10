package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/v2c/api"
	"github.com/docker/v2c/system"
	"github.com/docker/v2c/workflow"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"path/filepath"
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
					Name:  `rm`,
					Usage: `remove a transformed image`,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  `force, f`,
							Usage: `Remove any running containers that may be using IMAGE`,
						},
						cli.BoolFlag{
							Name:  `no-prune`,
							Usage: `Do no delete untagged parents`,
						},
					},
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

func buildHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Running image transformation.")

	abs, err := filepath.Abs(c.Args().Get(0))
	if err != nil {
		return err
	}
	_, err = os.Stat(abs)
	if err != nil {
		return err
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithCancel(context.Background())

	go func(cancel context.CancelFunc) {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}(cancel)

	_, err = workflow.Build(ctx, abs)
	return err
}

// management handlers

func installDetectiveHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Installing a detective")
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
	gone, ie := system.RemoveProducts(c.Args(), c.Bool(`force`), !c.Bool(`no-prune`))
	return renderRemoved(os.Stdout, gone, ie)
}

func exportImageHandler(c *cli.Context) error {
	if c.NArg() != 1 {
		return errExactlyOne
	}
	fmt.Println("Exporting an image")
	return nil
}

// list handlers

func listImageHandler(c *cli.Context) error {
	if c.Args().Present() {
		return errExactlyNone
	}
	products, err := system.ListProducedImages()
	if err != nil {
		return err
	}

	return renderTabbed(`imageList`, os.Stdout, struct {
		Products []api.Product
	}{
		Products: products,
	},
	)
}

func listProvisionerHandler(c *cli.Context) error {
	if c.NArg() > 0 {
		return errExactlyNone
	}
	components, err := system.DetectComponents()
	if err != nil {
		return err
	}

	return renderTabbed(`provisionerList`, os.Stdout, components)
}

func listDetectiveHandler(c *cli.Context) error {
	if c.NArg() > 0 {
		return errExactlyNone
	}

	components, err := system.DetectComponents()
	if err != nil {
		return err
	}

	return renderTabbed(`detectiveList`, os.Stdout, components)
}
