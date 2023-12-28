package model

// Tag is the tag for a bookmark.
type Tag struct {
	ID         int    `db:"id"          json:"id"`
	Name       string `db:"name"        json:"name"`
	NBookmarks int    `db:"n_bookmarks" json:"nBookmarks,omitempty"`
	Deleted    bool   `json:"-"`
}
