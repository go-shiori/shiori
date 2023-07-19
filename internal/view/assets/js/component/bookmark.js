var template = `
<div class="bookmark" :class="{list: listMode, 'no-thumbnail': hideThumbnail, selected: selected}">
	<a class="bookmark-selector" 
		v-if="editMode" 
		@click="selectBookmark">
	</a>
	<a class="bookmark-link" :href="mainURL" target="_blank" rel="noopener">
		<span class="thumbnail" v-if="thumbnailVisible" :style="thumbnailStyleURL"></span>
		<p class="title" dir="auto">{{title}}
			<i v-if="hasContent" class="fas fa-file-alt"></i>
			<i v-if="hasArchive" class="fas fa-archive"></i>
			<i v-if="public" class="fas fa-eye"></i>
		</p>
		<p class="excerpt" v-if="excerptVisible">{{excerpt}}</p>
		<p class="id" v-show="showId">{{id}}</p>
	</a>
	<div class="bookmark-tags" v-if="tags.length > 0">
		<a v-for="tag in tags" @click="tagClicked($event, tag.name)">{{tag.name}}</a>
	</div>
	<div class="spacer"></div>
	<div class="bookmark-menu">
		<a class="url" :href="url" target="_blank" rel="noopener">
			{{hostnameURL}}
		</a>
		<template v-if="!editMode && menuVisible">
			<a title="Edit bookmark" @click="editBookmark">
				<i class="fas fa-fw fa-pencil-alt"></i>
			</a>
			<a title="Delete bookmark" @click="deleteBookmark">
				<i class="fas fa-fw fa-trash-alt"></i>
			</a>
			<a title="Update archive" @click="updateBookmark">
				<i class="fas fa-fw fa-cloud-download-alt"></i>
			</a>
            <a v-if="hasEbook" title="Download book" @click="downloadebook">
                <i class="fas fa-fw fa-book"></i>
            </a>
		</template>
	</div>
</div>`;

export default {
	template: template,
	props: {
		id: Number,
		url: String,
		title: String,
		excerpt: String,
		public: Number,
		imageURL: String,
		hasContent: Boolean,
		hasArchive: Boolean,
		hasEbook: Boolean,
		index: Number,
		showId: Boolean,
		editMode: Boolean,
		listMode: Boolean,
		hideThumbnail: Boolean,
		hideExcerpt: Boolean,
		selected: Boolean,
		menuVisible: Boolean,
		tags: {
			type: Array,
			default() {
				return []
			}
		}
	},
	computed: {
		mainURL() {
			if (this.hasContent) {
				return new URL(`bookmark/${this.id}/content`, document.baseURI);
			} else if (this.hasArchive) {
				return new URL(`bookmark/${this.id}/archive`, document.baseURI);
			} else {
				return this.url;
			}
		},
		ebookURL() {
			if (this.hasEbook) {
				return new URL(`bookmark/${this.id}/ebook`, document.baseURI);
			} else  {
                return null;
            }
		},
		hostnameURL() {
			var url = new URL(this.url);
			return url.hostname.replace(/^www\./, "");
		},
		thumbnailVisible() {
			return this.imageURL !== "" &&
				!this.hideThumbnail;
		},
		excerptVisible() {
			return this.excerpt !== "" &&
				!this.thumbnailVisible &&
				!this.hideExcerpt;
		},
		thumbnailStyleURL() {
			return {
				backgroundImage: `url("${this.imageURL}")`
			}
		},
		eventItem() {
			return {
				id: this.id,
				index: this.index,
			}
		}
	},
	methods: {
		tagClicked(name, event) {
			this.$emit("tag-clicked", name, event);
		},
		selectBookmark() {
			this.$emit("select", this.eventItem);
		},
		editBookmark() {
			this.$emit("edit", this.eventItem);
		},
		deleteBookmark() {
			this.$emit("delete", this.eventItem);
		},
		updateBookmark() {
			this.$emit("update", this.eventItem);
		},
		downloadebook() {
            const id = this.id;
            const ebook_url = new URL(`bookmark/${id}/ebook`, document.baseURI);
            const downloadLink = document.createElement("a");
            downloadLink.href = ebook_url.toString();
            downloadLink.download = `${this.title}.epub`;
            downloadLink.click();
		},
	}
}
