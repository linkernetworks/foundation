package config

//DataConfig
type DataConfig struct {
	ImageDir        string `json:"imageDir"`
	ThumbnailDir    string `json:"thumbnailDir"`
	WorkspaceDir    string `json:"batchDir"`
	ArchiveDir      string `json:"archiveDir"`
	ModelDir        string `json:"modelDir"`
	ModelArchiveDir string `json:"modelArchiveDir"`
}
