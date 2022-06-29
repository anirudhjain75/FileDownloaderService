package models

import (
	"github.com/google/uuid"
	"net/http"
)

type FileDownloader struct {
	Id         uuid.UUID
	Downloads  map[uuid.UUID]Download
	ServeMux   *http.ServeMux
	Preference Preference
}

type Preference struct {
	IsSeq            bool
	DownloadLocation string
}

type Download struct {
	Id               uuid.UUID `json:"id,omitempty"`
	ListOfFiles      []string  `json:"listOfFiles"`
	IsSeq            bool      `json:"isSeq"`
	DownloadLocation string    `json:"downloadLocation,omitempty"`
}

type DownloadRequest struct {
	ListOfFiles      []string `json:"listOfFiles"`
	IsSeq            *bool    `json:"isSeq,omitempty"`
	DownloadLocation string   `json:"downloadLocation,omitempty"`
}

type DownloadedFileAndError struct {
	File  string
	Error error
}

type UpdatePreferenceRequest struct {
	IsSeq            *bool  `json:"isSeq,omitempty"`
	DownloadLocation string `json:"downloadLocation,omitempty"`
}
