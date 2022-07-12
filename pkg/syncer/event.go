package syncer

type Operation string

const (
	Created  = "created"
	Modified = "modified"
	Moved    = "moved"
	Deleted  = "deleted"
	Unknown  = "unknown"
)

type Event struct {
	Operation  Operation
	Name       string
	Path       string
	ParentPath string
}
