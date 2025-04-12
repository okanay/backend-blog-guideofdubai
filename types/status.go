package types

type Status string

const (
	Draft     Status = "draft"
	Published Status = "published"
	Archived  Status = "archived"
	Deleted   Status = "deleted"
)
