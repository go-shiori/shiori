var template = `
<div class="bookmark" :class="{list: listMode}">
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
        <a title="Edit bookmark" @click="editBookmark">
            <i class="fas fa-pencil-alt"></i>
        </a>
        <a title="Delete bookmark" @click="deleteBookmark">
            <i class="fas fa-trash-alt"></i>
        </a>
        <a title="Update cache" @click="updateBookmark">
            <i class="fas fa-cloud-download-alt"></i>
        </a>
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
        showId: Boolean,
        listMode: Boolean,
        tags: {
            type: Array,
            default () {
                return []
            }
        }
    },
    data() {
        return {};
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
        }
    },
    methods: {
        tagClicked(name) {
            this.$emit("tagClicked", name);
        },
        editBookmark() {
            this.$emit("edit", this.id);
        },
        deleteBookmark() {
            this.$emit("delete", this.id);
        },
        updateBookmark() {
            this.$emit("update", this.id);
        }
    }
}