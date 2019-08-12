var template = `
<div class="bookmark" :class="{list: listMode, selected: selected}">
    <a class="bookmark-selector" 
        v-if="editMode" 
        @click="selectBookmark">
    </a>
    <a class="bookmark-link" :href="mainURL" target="_blank" rel="noopener">
        <span class="thumbnail" v-if="imageURL" :style="thumbnailStyleURL"></span>
        <p class="title">{{title}}
            <i v-if="hasContent" class="fas fa-file-alt"></i>
            <i v-if="hasArchive" class="fas fa-archive"></i>
            <i v-if="public" class="fas fa-eye"></i>
        </p>
        <p class="excerpt" v-if="!imageURL">{{excerpt}}</p>
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
        <template v-if="!editMode">
            <a title="Edit bookmark" @click="editBookmark">
                <i class="fas fa-fw fa-pencil-alt"></i>
            </a>
            <a title="Delete bookmark" @click="deleteBookmark">
                <i class="fas fa-fw fa-trash-alt"></i>
            </a>
            <a title="Update archive" @click="updateBookmark">
                <i class="fas fa-fw fa-cloud-download-alt"></i>
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
        index: Number,
        showId: Boolean,
        editMode: Boolean,
        listMode: Boolean,
        selected: Boolean,
        tags: {
            type: Array,
            default () {
                return []
            }
        }
    },
    computed: {
        mainURL() {
            if (this.hasContent) return `/bookmark/${this.id}/content`;
            else if (this.hasArchive) return `/bookmark/${this.id}/archive`;
            else return this.url;
        },
        hostnameURL() {
            var url = new URL(this.url);
            return url.hostname.replace(/^www\./, "");
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
        }
    }
}