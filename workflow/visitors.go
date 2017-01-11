package workflow

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/docker/docker/builder/dockerfile/parser"
)

func addProductMetadata() error {
	b := new(bytes.Buffer)
	b.WriteString("LABEL com.docker.v2c.product=1\n")
	return appendDockerfile(b)
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
	if len(root.Children) <= 0 {
		return nil
	}

	if bad := verifyContributedInstructionsForCategory(`os`, root); bad != `` {
		return errors.New(fmt.Sprintf(`Provisioners in the OS category may only contribute a single FROM instruction. Illegal instructions: %v`, bad))
	}

	// Add an extra newline
	dfb := bytes.NewBuffer(df)
	dfb.WriteString("\n")
	return appendDockerfile(dfb)

}

func applyCategory(c string, ms []manifest) error {

	// This visitor is going to take all of the tars in the category and add ADD instructions for them to /
	var b bytes.Buffer
	for _, m := range ms {
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
				if bad := verifyContributedInstructionsForCategory(c, root); bad != `` {
					return errors.New(fmt.Sprintf("Illegal instruction in %v category Dockerfile fragment: %v contributed by %v:%v", c, bad, m.Provisioner.Repository, m.Provisioner.Tag))
				}
				b.Write(df)
				b.WriteString("\n")

			}
		}
	}
	return appendDockerfile(&b)
}

// detects illegal Dockerfile contributions by category
func verifyContributedInstructionsForCategory(c string, root *parser.Node) string {
	for _, child := range root.Children {
		switch c {
		case `os`:
			// only allow from
			if child.Value != `from` {
				return child.Value
			}
		case `application`:
			// allow anything but these
			if child.Value == `from` ||
				child.Value == `add` ||
				child.Value == `copy` ||
				child.Value == `shell` ||
				child.Value == `entrypoint` ||
				child.Value == `cmd` ||
				child.Value == `onbuild` ||
				child.Value == `stopsignal` ||
				child.Value == `maintainer` ||
				child.Value == `expose` ||
				child.Value == `healthcheck` {
				return child.Value
			}
		case `config`:
			// allow anything but these
			if child.Value == `from` ||
				child.Value == `add` ||
				child.Value == `copy` ||
				child.Value == `shell` ||
				child.Value == `entrypoint` ||
				child.Value == `cmd` ||
				child.Value == `onbuild` ||
				child.Value == `stopsignal` ||
				child.Value == `maintainer` ||
				child.Value == `healthcheck` {
				return child.Value
			}
		case `init`:
			// only allow these two
			if child.Value != `entrypoint` &&
				child.Value != `cmd` {
				return child.Value
			}
		}
	}
	return ``
}
