package docker

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/docker/api/types/container"
	"github.com/gitzup/agent/pkg/assets"
	"github.com/go-errors/errors"
)

func CreateJsonResultParser(path string, schema *assets.Schema, response interface{}) ContainerRunHandler {
	return func(ctx context.Context, c container.ContainerCreateCreatedBody) error {

		// read JSON response from resource (fail if missing or invalid)
		reader, _, err := cli.CopyFromContainer(ctx, c.ID, path)
		if err != nil {
			return errors.WrapPrefix(err, fmt.Sprintf("failed copying '%s' from container", path), 0)
		}
		//noinspection GoUnhandledErrorResult
		defer reader.Close()
		tr := tar.NewReader(reader)
		_, err = tr.Next()
		if err == io.EOF {
			return errors.WrapPrefix(err, fmt.Sprintf("could not find '%s' in container", path), 0)
		} else if err != nil {
			return errors.WrapPrefix(err, fmt.Sprintf("failed reading '%s' in container", path), 0)
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			return errors.WrapPrefix(err, fmt.Sprintf("failed reading '%s' in container", path), 0)
		}
		err = schema.ParseAndValidate(response, b)
		if err != nil {
			return errors.WrapPrefix(err, fmt.Sprintf("response from container is illegal"), 0)
		}

		return nil
	}
}
