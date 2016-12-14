package system

import (
	"github.com/docker/v2c/api"
	// "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

type Components struct {
	Detectives   []api.Detective
	Provisioners []api.Provisioner
	Packagers    []api.Packager
}

func DetectComponents() (Components, error) {
	_, err := docker.NewEnvClient()
	if err != nil {
		return Components{}, err
	}

	return Components{
		Detectives: []api.Detective{
			api.Detective{
				Repository: `repo1`,
				Tag:        `tag1`,
				Category:   `original`,
				Description: `first description`,
			},
		},
		Provisioners: []api.Provisioner{
			api.Provisioner{
				Repository: `repo2`,
				Tag:        `tag2`,
				Category:   `original`,
				Description: `second description`,
			},
		},
	}, nil
}

func ListProducedImages() ([]api.Product, error) {
	_, err := docker.NewEnvClient()
	if err != nil {
		return []api.Product{}, err
	}
	return []api.Product{
		api.Product{
			Repository: `local/product1`,
			Tag:        `001`,
			Original:   `product.vmdk`,
			Created:    `12/13/2017`,
		},
	}, nil
}
