package workflow

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/docker/v2c/api"
	"io"
	"io/ioutil"
	"os"
	path "path/filepath"
)

type manifest struct {
	Provisioner api.Provisioner
	TarballName string
}

type readableResults struct {
	Provisioner api.Provisioner
	Tarball     io.ReadCloser
}

func fetchContributedDockerfile(m manifest) ([]byte, error) {
	if m.TarballName == `` {
		return []byte{}, nil
	}
	dn, _, err := cwdAndPerms()
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path.Join(dn, m.Provisioner.Category, m.TarballName), os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := tar.NewReader(f)
	for {
		h, err := r.Next()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if h.Name != `Dockerfile` {
			continue
		}
		b := new(bytes.Buffer)
		if _, err := io.Copy(b, r); err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}
}

func isCWDEmpty() (bool, error) {
	dn, err := os.Getwd()
	if err != nil {
		return false, err
	}
	f, err := os.Open(dn)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func appendDockerfile(b *bytes.Buffer) error {
	dn, p, err := cwdAndPerms()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path.Join(dn, "Dockerfile"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, p)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(b.Bytes()); err != nil {
		return err
	}
	return nil
}

func cwdAndPerms() (string, os.FileMode, error) {
	d, err := os.Getwd()
	if err != nil {
		return ``, 0, err
	}
	di, err := os.Stat(d)
	if err != nil {
		return ``, 0, err
	}
	return d, di.Mode().Perm(), nil
}

// TODO: a better implementation might write the manifests and tarballs to disk as they are created by the provisioners
func persistProvisionerResults(r map[string][]provisionerResponse) (map[string][]manifest, error) {
	// get CWD and FileMode
	d, cwdPerm, err := cwdAndPerms()
	if err != nil {
		return nil, err
	}

	manifests := map[string][]manifest{}

	for c, prs := range r {
		// make a category subdir
		p := path.Join(d, c)
		_, err = os.Stat(p)
		if err != nil {
			if err = os.Mkdir(p, cwdPerm); err != nil {
				return nil, err
			}
		}
		err = nil
		// write each tarball and a brief JSON manifest for the response
		for _, pr := range prs {
			h := sha256.Sum256([]byte(fmt.Sprintf("%v:%v", pr.Provisioner.Repository, pr.Provisioner.Tag)))
			tbn := fmt.Sprintf("%x.tar", h)
			mn := fmt.Sprintf("%x.manifest", h)

			m := manifest{
				Provisioner: pr.Provisioner,
				TarballName: tbn,
			}
			mb, err := json.Marshal(m)
			if err != nil {
				return nil, err
			}
			manifests[c] = append(manifests[c], m)
			if err = ioutil.WriteFile(path.Join(p, mn), mb, cwdPerm); err != nil {
				return nil, err
			}
			if err = ioutil.WriteFile(path.Join(p, tbn), pr.Tarball.Bytes(), cwdPerm); err != nil {
				return nil, err
			}
		}
	}
	return manifests, nil
}
