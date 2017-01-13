package system

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"context"
	"strings"
	"github.com/docker/v2c/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	gcontext "golang.org/x/net/context"
)

var (
	labels = map[string]string {
		`detective`: `detective`,
		`provisioner`: `provisioner`,
		`packager`: `packager`,
		`product`: `com.docker.v2c.product`,
		`component`: `com.docker.v2c.component`,
		`category`: `com.docker.v2c.component.category`,
		`description`: `com.docker.v2c.component.description`,
		`related`: `com.docker.v2c.component.rel`,
	}
)

type Components struct {
	Detectives   []api.Detective
	Provisioners []api.Provisioner
	Packagers    []api.Packager
}

func DetectComponents() (Components, error) {
	result := Components{}
	client, err := docker.NewEnvClient()
	if err != nil {
		return result, err
	}

	f := filters.NewArgs()
	f.Add(`label`, labels[`component`])

	components, err := client.ImageList(gcontext.Background(), types.ImageListOptions{ Filters: f })
	if err != nil {
		return result, err
	}
	for _, img := range components {
		if img.Labels[labels[`component`]] == labels[`detective`] {
			result.Detectives = append(result.Detectives, detectivesFromImageSummary(img)...)
		} else if img.Labels[labels[`component`]] == labels[`provisioner`] {
			result.Provisioners = append(result.Provisioners, provisionersFromImageSummary(img)...)
		} else if img.Labels[labels[`component`]] == labels[`packager`] {
			result.Packagers = append(result.Packagers, packagersFromImageSummary(img)...)
		} else {
			panic(`Unknown component type detected: ` + img.ID)
		}
	}
	return result, nil
}

func ListProducedImages() ([]api.Product, error) {
	result := []api.Product{}
	client, err := docker.NewEnvClient()
	if err != nil {
		return result, err
	}
	f := filters.NewArgs()
	f.Add(`label`, labels[`product`])

	imgs, err := client.ImageList(gcontext.Background(), types.ImageListOptions{
		Filters: f,
	})
	if err != nil {
		return result, err
	}
	for _, img := range imgs {
		if len(img.RepoTags) > 0 {
			for _, t := range img.RepoTags {
				p := strings.Split(t, `:`)
				result = append(result, api.Product{
					ImageID: img.ID,
					Repository: p[0],
					Tag: p[1],
					Original: `original`,
					Created: `date`,
				})
			}
		} else {
			result = append(result, api.Product{
				ImageID: img.ID,
				Original: `original`,
				Created: `date`,
			})
		}

	}

	return result, nil
}

func RemoveProducts(is []string, f bool, p bool) ([]types.ImageDelete, error) {
	result := []types.ImageDelete{}
	client, err := docker.NewEnvClient()
	if err != nil {
		return result, err
	}
	opt := types.ImageRemoveOptions{
		Force: f,
		PruneChildren: p,
	}

	for _, i := range is {
		dels, err := client.ImageRemove(gcontext.Background(), i, opt)
		if err != nil {
			return result, err
		}
		result = append(result, dels...)
	}
	return result, nil
}

func LaunchPackager(ctx context.Context, p api.Packager) (string, error) {
	client, err := docker.NewEnvClient()
	if err != nil {
		return ``, err
	}

	// Create
	createResult, err := client.ContainerCreate(gcontext.Background(),
		&container.Config{
			Image: fmt.Sprintf(`%v:%v`, p.Repository, p.Tag),
		},
		&container.HostConfig{
			NetworkMode: `none`,
		},
		&network.NetworkingConfig{

		},
		fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%v/%v", p.Repository, p.Tag)))),
	)
	if err != nil {
		return ``, err
	}

	// Run
	err = client.ContainerStart(gcontext.Background(),
		createResult.ID,
		types.ContainerStartOptions{

		},
	)
	if err != nil {
		return ``, err
	}

	// return container ID
	return createResult.ID, nil
}

func LaunchDetective(ctx context.Context, pc string, c chan *bytes.Buffer, d api.Detective) {
	client, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// Start a container from the image described by d
	// include Volumes From pc

	// Create
	createResult, err := client.ContainerCreate(gcontext.Background(),
		&container.Config{
			Image: fmt.Sprintf(`%v:%v`, d.Repository, d.Tag),
		},
		&container.HostConfig{
			VolumesFrom: []string{pc},
			NetworkMode: `none`,
		},
		&network.NetworkingConfig{

		},
		fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%v/%v", d.Repository, d.Tag)))),
	)
	if err != nil {
		panic(err)
	}

	// attach to the container
	attachment, err := client.ContainerAttach(gcontext.Background(), createResult.ID, types.ContainerAttachOptions{ Stdin: false, Stdout: true, Stream: true })
	if err != nil {
		panic(err)
	}
	defer attachment.Close()

	// Run
	err = client.ContainerStart(gcontext.Background(),
		createResult.ID,
		types.ContainerStartOptions{

		},
	)
	if err != nil {
		panic(err)
	}

	// Skip the first 8 bytes...
	// TODO: do something a bit more purposeful
	_, err = attachment.Reader.Discard(8)
	if err != nil {
		panic(err)
	}

	// Read from attachment until conn is closed or eof or whatever...
	stdout := new(bytes.Buffer)
	_, err = io.Copy(stdout, attachment.Reader)
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Wait for the container to stop
	code, err := client.ContainerWait(ctx, createResult.ID)
	if err != nil {
		panic(err)
	}
	if code != 0 {
		fmt.Printf("No results for %v:%v code: %v\n", d.Repository, d.Tag ,code)
		stdout = nil
	}

	// Cleanup the detective container
	err = RemoveContainer(ctx, createResult.ID)
	if err != nil {
		panic(err)
	}

	// Send the collected output on c
	// TODO: Do we need a timeout here? I think so...
	c <- stdout
}

func LaunchProvisioner(ctx context.Context, in *bytes.Buffer, c chan *bytes.Buffer, p api.Provisioner) {
	client, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// Start a container from the image described by p
	createResult, err := client.ContainerCreate(gcontext.Background(),
		&container.Config{
			Image: fmt.Sprintf(`%v:%v`, p.Repository, p.Tag),
			Tty:   false,
			OpenStdin: true,
			StdinOnce: true,
		},
		&container.HostConfig{

		},
		&network.NetworkingConfig{

		},
		fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%v/%v", p.Repository, p.Tag)))),
	)
	if err != nil {
		panic(err)
	}

	// attach to the container
	attachment, err := client.ContainerAttach(gcontext.Background(), createResult.ID, types.ContainerAttachOptions{ Stdin: true, Stdout: true, Stream: true })
	if err != nil {
		panic(err)
	}
	defer attachment.Close()

	// Run
	err = client.ContainerStart(gcontext.Background(),
		createResult.ID,
		types.ContainerStartOptions{

		},
	)
	if err != nil {
		panic(err)
	}

	// write the data from in to the con
	_, err = in.WriteTo(attachment.Conn)
	if err != nil {
		panic(err)
	}
	attachment.CloseWrite()

	// Skip the first 8 bytes...
	// TODO: do something a bit more purposeful
	_, err = attachment.Reader.Discard(8)
	if err != nil {
		panic(err)
	}

	// read from attachment until conn is closed or eof or whatever...
	stdout := new(bytes.Buffer)
	_, err = io.Copy(stdout, attachment.Reader)
	if err != nil && err != io.EOF {
		panic(err)
	}

	// Wait for the container to stop
	code, err := client.ContainerWait(ctx, createResult.ID)
	if err != nil {
		panic(err)
	}
	if code != 0 {
		return
	}

	// Cleanup the provisioner container
	err = RemoveContainer(ctx, createResult.ID)
	if err != nil {
		panic(err)
	}

	// Send the collected output on c
	// TODO: Do we need a timeout here? I think so...
	c <- stdout
}

func RemoveContainer(ctx context.Context, cid string) error {
	client, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	return client.ContainerRemove(gcontext.Background(), cid, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force: true,
	})
}

func detectivesFromImageSummary(i types.ImageSummary) []api.Detective {
	result := []api.Detective{}
	if len(i.RepoTags) > 0 {
		for _, t := range i.RepoTags {
			p := strings.Split(t, `:`)
			if len(p) < 2 {
				panic(`Malformed RepoTag for image ID: ` + i.ID)
			}
			result = append(result, api.Detective{
				ImageID: i.ID,
				Repository: p[0],
				Tag: p[1],
				Category: i.Labels[labels[`category`]],
				Description: i.Labels[labels[`description`]],
				Related: i.Labels[labels[`related`]],
			})
		}
	} else {
		result = append(result, api.Detective{
			ImageID: i.ID,
			Repository: `<none>`,
			Tag: `<none>`,
			Category: i.Labels[labels[`category`]],
			Description: i.Labels[labels[`description`]],
		})
	}
	return result
}

func provisionersFromImageSummary(i types.ImageSummary) []api.Provisioner {
	result := []api.Provisioner{}
	if len(i.RepoTags) > 0 {
		for _, t := range i.RepoTags {
			p := strings.Split(t, `:`)
			if len(p) < 2 {
				panic(`Malformed RepoTag for image ID: ` + i.ID)
			}
			result = append(result, api.Provisioner{
				ImageID: i.ID,
				Repository: p[0],
				Tag: p[1],
				Category: i.Labels[labels[`category`]],
				Description: i.Labels[labels[`description`]],
			})
		}
	} else {
		result = append(result, api.Provisioner{
			ImageID: i.ID,
			Repository: `<none>`,
			Tag: `<none>`,
			Category: i.Labels[labels[`category`]],
			Description: i.Labels[labels[`description`]],
		})
	}
	return result
}

func packagersFromImageSummary(i types.ImageSummary) []api.Packager {
	result := []api.Packager{}
	if len(i.RepoTags) > 0 {
		for _, t := range i.RepoTags {
			p := strings.Split(t, `:`)
			if len(p) < 2 {
				panic(`Malformed RepoTag for image ID: ` + i.ID)
			}
			result = append(result, api.Packager{
				ImageID: i.ID,
				Repository: p[0],
				Tag: p[1],
				Category: i.Labels[labels[`category`]],
				Description: i.Labels[labels[`description`]],
			})
		}
	} else {
		result = append(result, api.Packager{
			ImageID: i.ID,
			Repository: `<none>`,
			Tag: `<none>`,
			Category: i.Labels[labels[`category`]],
			Description: i.Labels[labels[`description`]],
		})
	}
	return result
}

