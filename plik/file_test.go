package plik

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/root-gg/plik/server/common"
)

func TestGetFileURL(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := startWithClient(ps, pc)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	data_byte := []byte(data)

	upload, file, err := pc.UploadReader("filename", bytes.NewBufferString(data))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Files(), 1, "invalid files count")

	fileURL, err := file.GetURL()
	require.NoError(t, err, "unable to get file URL")

	req, err := http.NewRequest("GET", fileURL.String(), &bytes.Buffer{})
	require.NoError(t, err, "unable to create request")

	resp, err := pc.HTTPClient.Do(req)
	require.NoError(t, err, "unable to execute request")
	require.Equal(t, http.StatusOK, resp.StatusCode, "invalid response status code")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "unable to read response body")
	require.Equal(t, data, string(body), "invalid file content")

	req, err = http.NewRequest("GET", fileURL.String(), &bytes.Buffer{})
    req.Header.Add("Range", "bytes=0-7")
	require.NoError(t, err, "unable to create request")

	resp, err = pc.HTTPClient.Do(req)
	require.NoError(t, err, "unable to execute request")
	require.Equal(t, http.StatusPartialContent, resp.StatusCode, "invalid response status code")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "unable to read response body")
    require.Equal(t, data_byte[0:8], body, "invalid file content")

	req, err = http.NewRequest("GET", fileURL.String(), &bytes.Buffer{})
    req.Header.Add("Range", "bytes=7-14")
	require.NoError(t, err, "unable to create request")

	resp, err = pc.HTTPClient.Do(req)
	require.NoError(t, err, "unable to execute request")
	require.Equal(t, http.StatusPartialContent, resp.StatusCode, "invalid response status code")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "unable to read response body")
    require.Equal(t, data_byte[7:], body, "invalid file content")
}

func TestNotUploadedGetFileURL(t *testing.T) {
	_, pc := newPlikServerAndClient()

	upload := pc.NewUpload()
	file := upload.AddFileFromReader("filename", bytes.NewBufferString("data"))

	_, err := file.GetURL()
	common.RequireError(t, err, "upload has not been created yet")

	upload.metadata = &common.Upload{}
	upload.metadata.InitializeForTests()

	_, err = file.GetURL()
	common.RequireError(t, err, "file has not been uploaded yet")
}

func TestFileWithURL(t *testing.T) {
	ps, pc := newPlikServerAndClient()
	defer ps.ShutdownNow()

	err := startWithClient(ps, pc)
	require.NoError(t, err, "unable to start plik server")

	data := "data data data"
	upload, file, err := pc.UploadReader("filename", bytes.NewBufferString(data))
	require.NoError(t, err, "unable to upload file")
	require.Len(t, upload.Files(), 1, "invalid files count")

	result := file.WithURL()
	require.NotNil(t, result, "WithURL should not return nil")
	require.Equal(t, file.Name, result.Name, "file name should match")
	require.NotEmpty(t, result.URL, "file URL should not be empty")
	require.Contains(t, result.URL, file.Metadata().ID, "file URL should contain file ID")
}

func TestFileWithURLBeforeUpload(t *testing.T) {
	_, pc := newPlikServerAndClient()

	upload := pc.NewUpload()
	file := upload.AddFileFromReader("test.txt", bytes.NewBufferString("data"))

	result := file.WithURL()
	require.NotNil(t, result, "WithURL should not return nil even before upload")
	require.Empty(t, result.URL, "URL should be empty before upload")
}
