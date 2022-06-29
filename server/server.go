package server

import (
	"FileDownloaderService/Utils"
	"FileDownloaderService/models"
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// DownloadHandler should handle both read and write. Read in GET req and Write is POST req
func DownloadHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		urlValues := r.URL.Query()

		downloadId := urlValues["downloadId"][0]
		if downloadId == "" {
			http.Error(w, "No Input Provided", http.StatusBadRequest)
			return
		}

		if v, ok := fileDownloader.Downloads[uuid.MustParse(downloadId)]; ok {
			bStr, err := json.Marshal(v)
			if err != nil {
				http.Error(w, "Error while compiling response", http.StatusInternalServerError)
				return
			}
			_, err = w.Write(bStr)
			if err != nil {
				http.Error(w, "Error while writing response", http.StatusInternalServerError)
			}
			return
		} else {
			_, err := w.Write([]byte("Unable to find download with id: " + downloadId))
			if err != nil {
				http.Error(w, "Error while writing response", http.StatusInternalServerError)
			}
			return
		}
	}

	download := models.DownloadRequest{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, &download)
	if err != nil {
		http.Error(w, "Invalid Request Payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if download.DownloadLocation == "" {
		download.DownloadLocation = fileDownloader.Preference.DownloadLocation
	}

	if download.IsSeq == nil {
		download.IsSeq = new(bool)
		*download.IsSeq = fileDownloader.Preference.IsSeq
	}

	if *download.IsSeq {

		files := make([]string, len(download.ListOfFiles))
		for i, v := range download.ListOfFiles {
			dfe := Utils.DownloadFile(v, download.DownloadLocation)
			if err != nil {
				http.Error(w, "error downloading file :"+v, http.StatusInternalServerError)
				return
			}
			files[i] = dfe.File
		}

		downloadId := uuid.New()

		fileDownloader.Downloads[downloadId] = models.Download{
			Id:               downloadId,
			ListOfFiles:      files,
			IsSeq:            *download.IsSeq,
			DownloadLocation: download.DownloadLocation,
		}

		_, err := w.Write([]byte(downloadId.String()))
		if err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}

	} else {

		files := make([]string, 0)

		output := make(chan models.DownloadedFileAndError)

		for _, v := range download.ListOfFiles {
			go func() {
				output <- Utils.DownloadFile(v, download.DownloadLocation)
			}()
		}

		for i := 0; i < len(download.ListOfFiles); i++ {
			dfe := <-output
			if dfe.Error != nil {
				http.Error(w, "Error downloading on file\n"+dfe.Error.Error()+"\n"+dfe.File+"\n", http.StatusInternalServerError)
				return
			}
			files = append(files, dfe.File)
		}

		downloadId := uuid.New()

		fileDownloader.Downloads[downloadId] = models.Download{
			Id:               downloadId,
			ListOfFiles:      files,
			IsSeq:            *download.IsSeq,
			DownloadLocation: download.DownloadLocation,
		}

		_, err := w.Write([]byte(downloadId.String()))
		if err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}

	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("ok"))
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func UpdatePreferenceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPatch:
		{
			updatePreference := models.UpdatePreferenceRequest{}

			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "error reading request body", http.StatusBadRequest)
				return
			}

			err = json.Unmarshal(data, &updatePreference)
			if err != nil {
				http.Error(w, "unable to parse request body", http.StatusInternalServerError)
				return
			}

			if updatePreference.IsSeq != nil {
				fileDownloader.Preference.IsSeq = *updatePreference.IsSeq
			}

			if updatePreference.DownloadLocation != "" {
				fileDownloader.Preference.DownloadLocation = updatePreference.DownloadLocation
			}

			_, err = w.Write([]byte("ok"))
			if err != nil {
				http.Error(w, "error while writing response: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

	case http.MethodGet:
		{

			data, err := json.Marshal(fileDownloader.Preference)
			if err != nil {
				http.Error(w, "unable to parse preferences"+err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = w.Write(data)
			if err != nil {
				http.Error(w, "error while writing response: "+err.Error(), http.StatusInternalServerError)
				return
			}

		}
	}
}

var fileDownloader models.FileDownloader

func Init() models.FileDownloader {

	m := http.NewServeMux()

	m.Handle("/download", http.TimeoutHandler(http.HandlerFunc(DownloadHandler), time.Second*2, "RequestTimeout"))
	m.Handle("/health", http.TimeoutHandler(http.HandlerFunc(HealthHandler), time.Second*2, "RequestTimeout"))
	m.Handle("/updatePreference", http.TimeoutHandler(http.HandlerFunc(UpdatePreferenceHandler), time.Second*2, "RequestTimeout"))

	downloads := make(map[uuid.UUID]models.Download)

	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	fileDownloader = models.FileDownloader{
		Id:        uuid.New(),
		Downloads: downloads,
		ServeMux:  m,
		Preference: models.Preference{
			IsSeq:            false,
			DownloadLocation: currentDirectory,
		},
	}

	err = http.ListenAndServe(":8080", fileDownloader.ServeMux)
	if err != nil {
		log.Fatalln("Unable to Initiate fileDownloader %v", err.Error())
	}

	return fileDownloader
}
