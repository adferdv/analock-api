package models

type InternetArchiveSearchResponse struct {
	Response InternetArchiveBookResponse `json:"response"`
}

type InternetArchiveBookResponse struct {
	NumFound int                   `json:"numFound"`
	Start    int                   `json:"start"`
	Docs     []InternetArchiveBook `json:"docs"`
}

type InternetArchiveBook struct {
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Creator    string `json:"creator"`
}

type InternetArchiveMetadataResponse struct {
	Files    []InternetArchiveFile   `json:"files"`
	Metadata InternetArchiveMetadata `json:"metadata"`
}

type InternetArchiveFile struct {
	Name   string `json:"name"`
	Format string `json:"format"`
}

type InternetArchiveMetadata struct {
	Identifier       string      `json:"identifier"`
	Mediatype        string      `json:"mediatype"`
	Collection       []string    `json:"collection"`
	Description      string      `json:"description"`
	Scanner          string      `json:"scanner"`
	Subject          interface{} `json:"subject"`
	Title            string      `json:"title"`
	Publicdate       string      `json:"publicdate"`
	Uploader         string      `json:"uploader"`
	Addeddate        string      `json:"addeddate"`
	Language         string      `json:"language"`
	IdentifierAccess string      `json:"identifier-access"`
	IdentifierArk    string      `json:"identifier-ark"`
	Ppi              string      `json:"ppi"`
	Ocr              string      `json:"ocr"`
	RepubState       string      `json:"repub_state"`
	Curation         string      `json:"curation"`
	BackupLocation   string      `json:"backup_location"`
}
