var template = `
<div class="bookmark" :class="{list: listMode, selected: selected}">
    <a class="bookmark-selector" 
        v-if="editMode" 
        @click="selectBookmark">
    </a>
    <a class="bookmark-link" :href="url" target="_blank">
        <span class="thumbnail" v-if="imageURL" :style="thumbnailStyleURL"></span>
        <p class="title">{{title}}</p>
        <p class="excerpt" v-if="!imageURL">{{excerpt}}</p>
        <p class="id" v-show="showId">{{id}}</p>
    </a>
    <div class="bookmark-tags" v-if="tags.length > 0">
        <a v-for="tag in tags" @click="tagClicked(tag.name)">{{tag.name}}</a>
    </div>
    <div class="spacer"></div>
    <div class="bookmark-menu">
        <a class="url" :href="url" target="_blank">
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
        imageURL: String,
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
        hostnameURL() {
            var url = new URL(this.url);
            return url.hostname;
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
        tagClicked(name) {
            this.$emit("tagClicked", name);
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