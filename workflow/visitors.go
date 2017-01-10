package workflow

import (
	"bytes"
	"errors"
	"fmt"
	path "path/filepath"
	"github.com/docker/docker/builder/dockerfile/parser"
)

func getAddDirectivesForTars(ms []manifest) ([]byte, error) {
	b := new(bytes.Buffer)
	for _, m := range ms {
		b.WriteString(fmt.Sprintf("ADD %v/%v /\n", path.Join(`.`, m.Provisioner.Category), m.TarballName))
	}
	return b.Bytes(), nil
}

func applyOSCategory(c []manifest) error {
	if len(c) > 1 {
		return errors.New(`OS category contains multiple results.`)
	}
	m := c[0]
	df, err := fetchContributedDockerfile(m)
	if err != nil {
		return err
	}
	if len(df) <= 0 {
		appendDockerfile(bytes.NewBufferString("FROM scratch"))
		return nil
	}
	dfr := bytes.NewReader(df)
	s := parser.Directive{}
	if err = parser.SetEscapeToken(parser.DefaultEscapeToken, &s); err != nil {
		return err
	}
	root, err := parser.Parse(dfr, &s)
	if err != nil {
		return err
	}
	if len(root.Children) != 1 || root.Children[0].Value != `from` {
		return errors.New(`Provisioners in the OS category may only contribute a single FROM instruction.`)
	}
	// Add an extra newline
	dfb := bytes.NewBuffer(df)
	dfb.WriteString("\n")
	return appendDockerfile(dfb)
	
}

func applyApplicationCategory(c []manifest) error {

	// This visitor is going to take all of the tars in the category and add ADD instructions for them to /
	var b bytes.Buffer
	for _, m := range c {
		// The ADD instruction unpacks the tar file at the root.
		// Only files with the exact same fully qualified name will be in conflict. 
		// This isn't a problem because all these files are being sourced from the same vmdk.
		// In this special case conflicts are simply redundant.
		b.WriteString(fmt.Sprintf("ADD ./%s/%s /\n", m.Provisioner.Category, m.TarballName))

		// Grab the contributed Dockerfile fragment, validate no illegal instructions, and append.
		df, err := fetchContributedDockerfile(m)
		if err != nil {
			return err
		}
		if len(df) > 0 {
			dfr := bytes.NewReader(df)
			s := parser.Directive{}
			if err = parser.SetEscapeToken(parser.DefaultEscapeToken, &s); err != nil {
				return err
			}
			root, err := parser.Parse(dfr, &s)
			if err != nil {
				return err
			}
			if len(root.Children) > 0 {
				for _, child := range root.Children {
					if child.Value == `from` || 
					   child.Value == `shell` || 
					   child.Value == `entrypoint` ||
					   child.Value == `cmd` ||
					   child.Value == `onbuild` ||
					   child.Value == `stopsignal` ||
					   child.Value == `maintainer` ||
					   child.Value == `healthcheck` {
						return errors.New(fmt.Sprintf("Illegal instruction in application category Dockerfile fragment: %v contributed by %v:%v", child.Value, m.Provisioner.Repository, m.Provisioner.Tag))
					}
				}
				b.Write(df)
				b.WriteString("\n")
				
			}
		}
	}
	return appendDockerfile(&b)
}
