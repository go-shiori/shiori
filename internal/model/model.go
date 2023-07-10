package model

// Tag is the tag for a bookmark.
type Tag struct {
	ID         int    `db:"id"          json:"id"`
	Name       string `db:"name"        json:"name"`
	NBookmarks int    `db:"n_bookmarks" json:"nBookmarks,omitempty"`
	Deleted    bool   `json:"-"`
}

// Bookmark is the record for an URL.
type Bookmark struct {
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
	HasArchive    bool   `json:"hasArchive"`
	HasEbook      bool   `json:"hasEbook"`
	Tags          []Tag  `json:"tags"`
	CreateArchive bool   `json:"createArchive"`
	CreateEbook   bool   `json:"createEbook"`
}

// Account is person that allowed to access web interface.
type Account struct {
	ID            int    `db:"id"                   json:"id"`
	Username      string `db:"username"             json:"username"`
	Password      string `db:"password"             json:"password,omitempty"`
	Owner         bool   `db:"owner"                json:"owner"`
	Showid        bool   `db:"showid"               json:"showid"`
	Listmode      bool   `db:"listmode"             json:"listmode"`
	Hidethumbnail bool   `db:"hidethumbnail"        json:"hidethumbnail"`
	Hideexcerpt   bool   `db:"hideexcerpt"          json:"hideexcerpt"`
	Nightmode     bool   `db:"nightmode"            json:"nightmode"`
	Keepmetadata  bool   `db:"keepmetadata"         json:"keepmetadata"`
	Usearchive    bool   `db:"usearchive"           json:"usearchive"`
	Makepublic    bool   `db:"makepublic"           json:"makepublic"`
}
