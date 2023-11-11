package model

// BookmarkDTO is the bookmark object representation in database and the data transfer object
// at the same time, pending a refactor to two separate object to represent each role.
type BookmarkDTO struct {
	ID            int    `db:"id"            json:"id"`
	URL           string `db:"url"           json:"url"`
	Title         string `db:"title"         json:"title"`
	Excerpt       string `db:"excerpt"       json:"excerpt"`
	Author        string `db:"author"        json:"author"`
	Public        int    `db:"public"        json:"public"`
	Modified      string `db:"modified"      json:"modified"`
	Content       string `db:"content"       json:"-"`
	HTML          string `db:"html"          json:"html,omitempty"`
	ImageURL      string `db:"image_url"     json:"imageURL"`
	HasContent    bool   `db:"has_content"   json:"hasContent"`
	Tags          []Tag  `json:"tags"`
	HasArchive    bool   `json:"hasArchive"`
	HasEbook      bool   `json:"hasEbook"`
	CreateArchive bool   `json:"create_archive"`
	CreateEbook   bool   `json:"create_ebook"`
}
