package Utils

import (
	"FileDownloaderService/models"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

/*

DownloadFile			— 	Download file
UpdatePreference		— 	Update Preference of Downloader
GetDownloadedFiles		— 	Get List of Files from Download ID
HealthCheck

*/

func DownloadFile(URL string, DownloadDir string) models.DownloadedFileAndError {

	dfe := models.DownloadedFileAndError{
		File:  "",
		Error: nil,
	}

	fileURL, err := url.Parse(URL)
	if err != nil {
		dfe.Error = err
		dfe.File = URL
		return dfe
	}

	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	err = os.Chdir(DownloadDir)
	if err != nil {
		dfe.Error = err
		dfe.File = URL
		return dfe
	}

	file, err := os.Create(fileName)
	if err != nil {
		dfe.Error = err
		dfe.File = URL
		return dfe
	}

	client := http.Client{}

	resp, err := client.Get(URL)
	if err != nil {
		dfe.Error = err
		dfe.File = URL
		return dfe
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		dfe.Error = err
		dfe.File = URL
		return dfe
	}

	newFileURL := DownloadDir + "/" + fileName

	defer file.Close()

	dfe.File = newFileURL

	return dfe
}
