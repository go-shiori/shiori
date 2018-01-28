package model

type Tag struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type Bookmark struct {
	ID          int64  `db:"id"`
	URL         string `db:"url"`
	Title       string `db:"title"`
	ImageURL    string `db:"image_url"`
	Excerpt     string `db:"excerpt"`
	Author      string `db:"author"`
	Language    string `db:"language"`
	MinReadTime int    `db:"min_read_time"`
	MaxReadTime int    `db:"max_read_time"`
	Modified    string `db:"modified"`
	Tags        []Tag
}
