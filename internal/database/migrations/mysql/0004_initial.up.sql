CREATE TABLE IF NOT EXISTS bookmark_tag(
		bookmark_id INT(11)      NOT NULL,
		tag_id      INT(11)      NOT NULL,
		PRIMARY KEY(bookmark_id, tag_id),
		KEY bookmark_tag_bookmark_id_FK (bookmark_id),
		KEY bookmark_tag_tag_id_FK (tag_id),
		CONSTRAINT bookmark_tag_bookmark_id_FK FOREIGN KEY (bookmark_id) REFERENCES bookmark (id),
		CONSTRAINT bookmark_tag_tag_id_FK FOREIGN KEY (tag_id) REFERENCES tag (id))
		CHARACTER SET utf8mb4;
