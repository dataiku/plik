package zip

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Backend config
type Backend struct {
	Config *BackendConfig
}

// NewZipBackend instantiate a new ZIP Archive Backend
// and configure it from config map
func NewZipBackend(config map[string]any) (zb *Backend, err error) {
	zb = new(Backend)
	zb.Config = NewZipBackendConfig(config)
	if _, err = os.Stat(zb.Config.Zip); os.IsNotExist(err) || os.IsPermission(err) {
		if zb.Config.Zip, err = exec.LookPath("zip"); err != nil {
			err = errors.New("zip binary not found in $PATH, please install or edit ~/.plikrc")
		}
	}
	return
}

// Configure implementation for ZIP Archive Backend
func (zb *Backend) Configure(arguments map[string]any) (err error) {
	if arguments["--archive-options"] != nil && arguments["--archive-options"].(string) != "" {
		zb.Config.Options = arguments["--archive-options"].(string)
	}
	return
}

// Archive implementation for ZIP Archive Backend
func (zb *Backend) Archive(files []string) (reader io.Reader, err error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("unable to make a zip archive from STDIN")
	}

	var args []string
	args = append(args, strings.Fields(zb.Config.Options)...)
	args = append(args, "-r", "-")
	args = append(args, files...)

	reader, writer := io.Pipe()

	var stderr bytes.Buffer
	cmd := exec.Command(zb.Config.Zip, args...)
	cmd.Stdout = writer
	cmd.Stderr = &stderr

	go func() {
		err := cmd.Start()
		if err != nil {
			_ = writer.CloseWithError(fmt.Errorf("unable to start zip cmd: %w", err))
			return
		}
		err = cmd.Wait()
		if err != nil {
			if stderr.Len() > 0 {
				_ = writer.CloseWithError(fmt.Errorf("zip cmd failed: %w: %s", err, stderr.String()))
			} else {
				_ = writer.CloseWithError(fmt.Errorf("zip cmd failed: %w", err))
			}
			return
		}
		_ = writer.Close()
	}()

	return reader, nil
}

// Comments implementation for ZIP Archive Backend
// Left empty because ZIP can't accept piping to it's STDIN
func (zb *Backend) Comments() string {
	return ""
}

// GetFileName returns the final archive file name
func (zb *Backend) GetFileName(files []string) (name string) {
	name = "archive"
	if len(files) == 1 {
		name = filepath.Base(files[0])
	}
	name += ".zip"
	return
}

// GetConfiguration implementation for ZIP Archive Backend
func (zb *Backend) GetConfiguration() any {
	return zb.Config
}
