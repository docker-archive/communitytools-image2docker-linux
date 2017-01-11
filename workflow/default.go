package workflow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/v2c/api"
	"github.com/docker/v2c/system"
)

var errNotYetImplemented = errors.New(`not yet implemented`)

type detectiveResponse struct {
	Category string
	Next     string
	Tarball  *bytes.Buffer
}

type provisionerResponse struct {
	Provisioner api.Provisioner
	Category    string
	Tarball     *bytes.Buffer
}

func Build(ctx context.Context, target string) (string, error) {
	components, err := system.DetectComponents()
	if err != nil {
		return ``, nil
	}
	if len(components.Packagers) == 0 {
		return ``, errors.New(`no installed packagers`)
	}

	// Choose a Packager
	packager := choosePackager(components)
	fmt.Printf("Using packager: %v:%v\n", packager.Repository, packager.Tag)

	// Launch the Packager
	pc, err := system.LaunchPackager(ctx, packager)
	if err != nil {
		return ``, err
	}
	defer system.RemoveContainer(ctx, pc)

	// Launch Detectives
	dr := make(chan detectiveResponse)
	for _, d := range components.Detectives {
		go launchDetective(ctx, d, pc, dr)
	}

	// Collect Detective responses
	detected := []detectiveResponse{}
	collectDetectiveResponses(ctx, len(components.Detectives), dr, &detected)

	pCount := len(detected)

	// Shutdown the Packager
	err = system.RemoveContainer(ctx, pc)
	if err != nil {
		return ``, err
	}

	// Should quit early?
	if pCount == 0 {
		return ``, errors.New(`No components were detected.`)
	}

	// Launch Provisioners
	prc := make(chan provisionerResponse)
	err = launchProvisioners(ctx, components, prc, &detected)
	if err != nil {
		return ``, err
	}

	// Collect provisioned build contexts
	results := map[string][]provisionerResponse{}
	collectProvisionerResponses(ctx, pCount, prc, results)

	//for category, prs := range results {
	//	for _, tres := range prs {
	//		fmt.Printf("Category: %v, Provisioner Tarball: %v\n\n", category, tres.Tarball)
	//	}
	//}

	// We can cache at this point and prompt for conflict resolution if required.
	// At this point we have a fully analyzed image and proposals for provisioning.
	ms, err := persistProvisionerResults(results)
	if err != nil {
		return ``, err
	}

	// Build context assembly pipeline
	// This could look like a pipeline where the result of one phase is piped to the next.
	// But we'd end up copying an amazing amount of data in memory and pipelines / nested functions
	// can be difficult to read. For the PoC we'll use a visitor pattern instead.
	// Need to process in category ordering - Operating System > Tooling > Platform > Application > Configuration

	if err = applyOSCategory(ms["os"]); err != nil {
		return ``, err
	}

	if err = addProductMetadata(); err != nil {
		return ``, err
	}

	if err = applyApplicationCategory(ms["application"]); err != nil {
		return ``, err
	}

	// Dockerfile assembly phases
	// FROM > COPY/ADD > EXPORT > ENV > RUN > SHELL/ENTRYPOINT/COMMAND

	// TODO: run docker build

	return ``, nil
}

//
// Workflow subroutines
//

func launchProvisioners(ctx context.Context, components system.Components, c chan provisionerResponse, rs *[]detectiveResponse) error {
	for _, r := range *rs {
		// TODO: replace the following with a lookup for the detectiveResponse.Next
		var p api.Provisioner
		for _, p = range components.Provisioners {
			if s := fmt.Sprintf("%v:%v", p.Repository, p.Tag); s == r.Next {
				break
			}
		}

		go launchProvisioner(ctx, p, r.Tarball, c)
	}
	return nil
}

func collectDetectiveResponses(ctx context.Context, c int, rc chan detectiveResponse, rs *[]detectiveResponse) error {
	if rs == nil {
		panic(`Nil response set passed to collectDetectiveResponses`)
	}
	for i := 0; i < c; i++ {
		select {
		case <-ctx.Done():
			return errors.New(`Task cancelled or late.`)
		case r := <-rc:
			if r.Tarball == nil {
				continue
			}
			*rs = append(*rs, r)
		}
	}
	return nil
}

func collectProvisionerResponses(ctx context.Context, c int, rc chan provisionerResponse, rs map[string][]provisionerResponse) error {
	if rs == nil {
		panic(`Nil response map passed to collectProvisionerResponses`)
	}
	for i := 0; i < c; i++ {
		select {
		case <-ctx.Done():
			return errors.New(`Task cancelled or late.`)
		case pr := <-rc:
			rs[pr.Category] = append(rs[pr.Category], pr)
		}
	}
	return nil
}

func choosePackager(c system.Components) api.Packager {
	return c.Packagers[0]
}

//
// launch control
//

func launchDetective(ctx context.Context, d api.Detective, pc string, drc chan detectiveResponse) {
	r := detectiveResponse{
		Category: d.Category,
		Next:     d.Related,
	}
	tbc := make(chan *bytes.Buffer)
	go system.LaunchDetective(ctx, pc, tbc, d)

	select {
	case r.Tarball = <-tbc:
	case <-ctx.Done():
	}
	close(tbc)

	select {
	case <-ctx.Done():
	case drc <- r:
	}
}

func launchProvisioner(ctx context.Context, p api.Provisioner, in *bytes.Buffer, prc chan provisionerResponse) {
	r := provisionerResponse{
		Provisioner: p,
		Category:    p.Category,
	}
	tbc := make(chan *bytes.Buffer)
	go system.LaunchProvisioner(ctx, in, tbc, p)

	select {
	case r.Tarball = <-tbc:
		prc <- r
	case <-ctx.Done():
	}
	close(tbc)
}
