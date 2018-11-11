package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	. "github.com/gitzup/agent/internal/logger"
	"github.com/go-errors/errors"
)

type dockerPullStatus struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error"`
	ProgressText   string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

func Pull(ctx context.Context, progress bool, image string) error {

	// list images
	imageListArgs := filters.NewArgs()
	imageListArgs.Add("reference", image)
	imageList, err := cli.ImageList(ctx, types.ImageListOptions{Filters: imageListArgs})
	if err != nil {
		return errors.WrapPrefix(err, "failed pulling image", 0)
	}

	// if it is already present, return
	if !strings.HasSuffix(image, ":latest") && len(imageList) > 0 {
		return nil
	}

	// pull it
	From(ctx).Infof("Pulling image '%s'...\n", image)
	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{All: true})
	if err != nil {
		return errors.WrapPrefix(err, "failed pulling image", 0)
	}
	//noinspection GoUnhandledErrorResult
	defer reader.Close()

	// if we're not using a TTY, ignore output and return
	if progress {
		err = trackProgress(reader)
		if err != nil {
			return errors.New(err)
		}
		return nil
	} else if _, err := io.Copy(ioutil.Discard, reader); err != nil {
		return errors.WrapPrefix(err, "failure occurred while discarding pull progress report", 0)
	} else {
		return nil
	}
}

func trackProgress(reader io.ReadCloser) error {
	pullEvents := json.NewDecoder(reader)
	var event *dockerPullStatus
	var progressMap = make(map[string]float32, 10)
	for {
		if err := pullEvents.Decode(&event); err != nil {
			if err == io.EOF {
				fmt.Print("\r\033[K\r")
				break
			}
			return errors.New(err)
		}
		time.Sleep(100 * time.Millisecond)

		progressMap[event.ID+":"+event.Status] = float32(event.ProgressDetail.Current) / float32(event.ProgressDetail.Total) * 100
		switch event.Status {
		case "Downloading":
			progressMap[event.ID+":Pulling fs layer"] = 100
		case "Download complete":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
		case "Extracting":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
			progressMap[event.ID+":Verifying Checksum"] = 100
			progressMap[event.ID+":Download complete"] = 100
		case "Pull complete":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
			progressMap[event.ID+":Verifying Checksum"] = 100
			progressMap[event.ID+":Download complete"] = 100
			progressMap[event.ID+":Extracting"] = 100
			progressMap[event.ID+":Pull complete"] = 100
		case "Digest":
			progressMap[event.ID+":Pulling fs layer"] = 100
			progressMap[event.ID+":Downloading"] = 100
			progressMap[event.ID+":Verifying Checksum"] = 100
			progressMap[event.ID+":Download complete"] = 100
			progressMap[event.ID+":Extracting"] = 100
			progressMap[event.ID+":Pull complete"] = 100
		}
		builder := strings.Builder{}
		for _, progress := range progressMap {
			if progress > 0 && progress < 100 {
				builder.WriteString(fmt.Sprintf("%d%% ", uint8(progress)))
			}
		}
		if builder.Len() > 0 {
			fmt.Print("\r" + builder.String() + "\033[K")
		}
	}
	return nil
}
