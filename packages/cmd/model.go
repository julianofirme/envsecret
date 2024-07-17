package cmd

type FileStatus string

var (
	StatusCreated   FileStatus = "created"
	StatusUpdated   FileStatus = "updated"
	StatusRemoved   FileStatus = "removed"
	StatusUnchanged FileStatus = "unchanged"
)

type fileMetadata struct {
	Path             string
	Status           FileStatus
	ModificationTime string
	GoToStaging      bool
}
