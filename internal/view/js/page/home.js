var template = `
<div id="page-home">
    <div class="page-header">
        <input type="text" placeholder="Search url, keyword or tags" v-model.trim="search" @focus="$event.target.select()" @keyup.enter="searchBookmarks"/>
        <a title="Refresh storage" @click="reloadData">
            <i class="fas fa-fw fa-sync-alt" :class="loading && 'fa-spin'"></i>
        </a>
        <a title="Add new bookmark" @click="showDialogAdd">
            <i class="fas fa-fw fa-plus-circle"></i>
        </a>
        <a title="Show tags">
            <i class="fas fa-fw fa-tags"></i>
        </a>
        <a title="Batch edit" @click="toggleEditMode">
            <i class="fas fa-fw fa-pencil-alt"></i>
        </a>
    </div>
    <div class="page-header" id="edit-box" v-if="editMode">
        <p>{{selection.length}} items selected</p>
        <a title="Delete bookmark" @click="showDialogDelete(selection)">
            <i class="fas fa-fw fa-trash-alt"></i>
        </a>
        <a title="Add tags" @click="showDialogAddTags(selection)">
            <i class="fas fa-fw fa-tags"></i>
        </a>
        <a title="Update archives" @click="showDialogUpdateArchive(selection)">
            <i class="fas fa-fw fa-cloud-download-alt"></i>
        </a>
        <a title="Cancel" @click="toggleEditMode">
            <i class="fas fa-fw fa-times"></i>
        </a>
    </div>
    <div id="bookmarks-grid" ref="bookmarksGrid" :class="{list: displayOptions.listMode}">
        <bookmark-item v-for="(book, index) in bookmarks" 
            v-bind="book" 
            :index="index"
            :key="book.id" 
            :editMode="editMode"
            :showId="displayOptions.showId"
            :listMode="displayOptions.listMode"
            :selected="isSelected(book.id)"
            @select="toggleSelection"
            @tagClicked="filterTag"
            @edit="showDialogEdit"
            @delete="showDialogDelete"
            @update="showDialogUpdateArchive">
        </bookmark-item>
        <div class="pagination-box" v-if="maxPage > 0">
            <p>Page</p>
            <input type="text" 
                placeholder="1" 
                :value="page" 
                @focus="$event.target.select()" 
                @keyup.enter="changePage($event.target.value)" 
                :disabled="editMode">
            <p>{{maxPage}}</p>
            <div class="spacer"></div>
            <template v-if="!editMode">
                <a v-if="page > 2" title="Go to first page" @click="changePage(1)">
                    <i class="fas fa-fw fa-angle-double-left"></i>
                </a>
                <a v-if="page > 1" title="Go to previous page" @click="changePage(page-1)">
                    <i class="fa fa-fw fa-angle-left"></i>
                </a>
                <a v-if="page < maxPage" title="Go to next page" @click="changePage(page+1)">
                    <i class="fa fa-fw fa-angle-right"></i>
                </a>
                <a v-if="page < maxPage - 1" title="Go to last page" @click="changePage(maxPage)">
                    <i class="fas fa-fw fa-angle-double-right"></i>
                </a>
            </template>
        </div>
    </div>
    <p class="empty-message" v-if="!loading && listIsEmpty">No saved bookmarks yet :(</p>
    <div class="loading-overlay" v-if="loading"><i class="fas fa-fw fa-spin fa-spinner"></i></div>
    <custom-dialog v-bind="dialog"/>
</div>`

import bookmarkItem from "../component/bookmark.js";
import customDialog from "../component/dialog.js";
import basePage from "./base.js";

export default {
    template: template,
    mixins: [basePage],
    components: {
        bookmarkItem,
        customDialog
    },
    data() {
        return {
            loading: false,
            editMode: false,
            selection: [],

            search: "",
            page: 0,
            maxPage: 0,
            bookmarks: [],
            tags: [],
        }
    },
    computed: {
        listIsEmpty() {
            return this.bookmarks.length <= 0;
        }
    },
    methods: {
        reloadData() {
            if (this.loading) return;
            this.page = 1;
            this.search = "";
            this.loadData(true, true);
        },
        loadData(saveState, fetchTags) {
            if (this.loading) return;

            // Set default args
            saveState = (typeof saveState === "boolean") ? saveState : true;
            fetchTags = (typeof fetchTags === "boolean") ? fetchTags : false;

            // Parse search query
            var rxTagA = /['"]#([^'"]+)['"]/g, // "#tag with space"
                rxTagB = /(^|\s+)#(\S+)/g, // #tag-without-space
                keyword = this.search,
                tags = [],
                rxResult;

            // Fetch tag A first
            while (rxResult = rxTagA.exec(keyword)) {
                tags.push(rxResult[1]);
            }

            // Clear tag A from keyword
            keyword = keyword.replace(rxTagA, "");

            // Fetch tag B
            while (rxResult = rxTagB.exec(keyword)) {
                tags.push(rxResult[2]);
            }

            // Clear tag B from keyword, then trim keyword
            keyword = keyword.replace(rxTagB, "").trim().replace(/\s+/g, " ");

            // Prepare URL for API
            var url = new URL("/api/bookmarks", document.URL);
            url.search = new URLSearchParams({
                keyword: keyword,
                tags: tags.join(","),
                page: this.page
            });

            // Fetch data from API
            var skipFetchTags = Error("skip fetching tags");

            this.loading = true;
            fetch(url)
                .then(response => {
                    if (!response.ok) throw response;
                    return response.json();
                })
                .then(json => {
                    // Set data
                    this.page = json.page;
                    this.maxPage = json.maxPage;
                    this.bookmarks = json.bookmarks;

                    // Save state and change URL if needed
                    if (saveState) {
                        var history = {
                            activePage: "page-home",
                            search: this.search,
                            page: this.page
                        };

                        var urlQueries = [];
                        if (this.page > 1) urlQueries.push(`page=${this.page}`);
                        if (this.search !== "") urlQueries.push(`search=${this.search}`);

                        var url = "#home"
                        if (urlQueries.length > 0) {
                            url += `?${urlQueries.join("&")}`;
                        }

                        window.history.pushState(history, "page-home", url);
                    }

                    // Fetch tags if requested
                    if (fetchTags) {
                        return fetch("/api/tags");
                    } else {
                        this.loading = false;
                        throw skipFetchTags;
                    }
                })
                .then(response => {
                    if (!response.ok) throw response;
                    return response.json();
                })
                .then(json => {
                    this.tags = json;
                    this.loading = false;
                })
                .catch(err => {
                    this.loading = false;

                    if (err !== skipFetchTags) {
                        err.text().then(msg => {
                            this.showErrorDialog(`${msg} (${err.status})`);
                        })
                    }
                });
        },
        searchBookmarks() {
            this.page = 1;
            this.loadData();
        },
        changePage(page) {
            page = parseInt(page, 10) || 0;
            if (page >= this.maxPage) this.page = this.maxPage;
            else if (page <= 1) this.page = 1;
            else this.page = page;

            this.$refs.bookmarksGrid.scrollTop = 0;
            this.loadData();
        },
        filterTag(tagName) {
            var rxSpace = /\s+/g,
                newTag = rxSpace.test(tagName) ? `"#${tagName}"` : `#${tagName}`;

            if (!this.search.includes(newTag)) {
                this.search += ` ${newTag}`;
                this.loadData();
            }
        },
        toggleEditMode() {
            this.selection = [];
            this.editMode = !this.editMode;
        },
        toggleSelection(item) {
            var idx = this.selection.findIndex(el => el.id === item.id);
            if (idx === -1) this.selection.push(item);
            else this.selection.splice(idx, 1);
        },
        isSelected(bookId) {
            return this.selection.findIndex(el => el.id === bookId) > -1;
        },
        showDialogAdd() {
            this.showDialog({
                title: "New Bookmark",
                content: "Create a new bookmark",
                fields: [{
                    name: "url",
                    label: "Url, start with http://...",
                }, {
                    name: "title",
                    label: "Custom title (optional)"
                }, {
                    name: "excerpt",
                    label: "Custom excerpt (optional)",
                    type: "area"
                }, {
                    name: "tags",
                    label: "Comma separated tags (optional)",
                    separator: ",",
                    dictionary: this.tags.map(tag => tag.name)
                }],
                mainText: "OK",
                secondText: "Cancel",
                mainClick: (data) => {
                    // Make sure URL is not empty
                    if (data.url.trim() === "") {
                        this.showErrorDialog("URL must not empty");
                        return;
                    }

                    // Prepare tags
                    var tags = data.tags
                        .toLowerCase()
                        .replace(/\s+/g, " ")
                        .split(/\s*,\s*/g)
                        .filter(tag => tag.trim() !== "")
                        .map(tag => {
                            return {
                                name: tag.trim()
                            };
                        });

                    // Send data
                    var data = {
                        url: data.url.trim(),
                        title: data.title.trim(),
                        excerpt: data.excerpt.trim(),
                        tags: tags
                    };

                    this.dialog.loading = true;
                    fetch("/api/bookmarks", {
                            method: "post",
                            body: JSON.stringify(data),
                            headers: {
                                "Content-Type": "application/json",
                            },
                        })
                        .then(response => {
                            if (!response.ok) throw response;
                            return response.json();
                        })
                        .then(json => {
                            this.dialog.loading = false;
                            this.dialog.visible = false;
                            this.bookmarks.splice(0, 0, json);
                        })
                        .catch(err => {
                            this.dialog.loading = false;
                            err.text().then(msg => {
                                this.showErrorDialog(`${msg} (${err.status})`);
                            })
                        });
                }
            });
        },
        showDialogEdit(item) {
            // Check the item
            if (typeof item !== "object") return;

            var id = (typeof item.id === "number") ? item.id : 0,
                index = (typeof item.index === "number") ? item.index : -1;

            if (id < 1 || index < 0) return;

            // Get the existing bookmark value
            var book = JSON.parse(JSON.stringify(this.bookmarks[index])),
                strTags = book.tags.map(tag => tag.name).join(", ");

            this.showDialog({
                title: "Edit Bookmark",
                content: "Edit the bookmark's data",
                showLabel: true,
                fields: [{
                    name: "title",
                    label: "Title",
                    value: book.title,
                }, {
                    name: "excerpt",
                    label: "Excerpt",
                    type: "area",
                    value: book.excerpt,
                }, {
                    name: "tags",
                    label: "Tags",
                    value: strTags,
                    separator: ",",
                    dictionary: this.tags.map(tag => tag.name)
                }],
                mainText: "OK",
                secondText: "Cancel",
                mainClick: (data) => {
                    // Validate input
                    if (data.title.trim() === "") return;

                    // Prepare tags
                    var tags = data.tags
                        .toLowerCase()
                        .replace(/\s+/g, " ")
                        .split(/\s*,\s*/g)
                        .filter(tag => tag.trim() !== "")
                        .map(tag => {
                            return {
                                name: tag.trim()
                            };
                        });

                    // Set new data
                    book.title = data.title.trim();
                    book.excerpt = data.excerpt.trim();
                    book.tags = tags;

                    // Send data
                    this.dialog.loading = true;
                    fetch("/api/bookmarks", {
                            method: "put",
                            body: JSON.stringify(book),
                            headers: {
                                "Content-Type": "application/json",
                            },
                        })
                        .then(response => {
                            if (!response.ok) throw response;
                            return response.json();
                        })
                        .then(json => {
                            this.dialog.loading = false;
                            this.dialog.visible = false;
                            this.bookmarks.splice(index, 1, json);
                        })
                        .catch(err => {
                            this.dialog.loading = false;
                            err.text().then(msg => {
                                this.showErrorDialog(`${msg} (${err.status})`);
                            })
                        });
                }
            });
        },
        showDialogDelete(items) {
            // Check and filter items
            if (typeof items !== "object") return;
            if (!Array.isArray(items)) items = [items];

            items = items.filter(item => {
                var id = (typeof item.id === "number") ? item.id : 0,
                    index = (typeof item.index === "number") ? item.index : -1;

                return id > 0 && index > -1;
            });

            if (items.length === 0) return;

            // Split ids and indices
            var ids = items.map(item => item.id),
                indices = items.map(item => item.index).sort((a, b) => b - a);

            // Create title and content
            var title = "Delete Bookmarks",
                content = "Delete the selected bookmarks ? This action is irreversible.";

            if (items.length === 1) {
                title = "Delete Bookmark";
                content = "Are you sure ? This action is irreversible.";
            }

            // Show dialog
            this.showDialog({
                title: title,
                content: content,
                mainText: "Yes",
                secondText: "No",
                mainClick: () => {
                    this.dialog.loading = true;
                    fetch("/api/bookmarks", {
                            method: "delete",
                            body: JSON.stringify(ids),
                            headers: {
                                "Content-Type": "application/json",
                            },
                        })
                        .then(response => {
                            if (!response.ok) throw response;
                            return response;
                        })
                        .then(() => {
                            this.selection = [];
                            this.editMode = false;
                            this.dialog.loading = false;
                            this.dialog.visible = false;
                            indices.forEach(index => this.bookmarks.splice(index, 1))

                            if (this.bookmarks.length < 20) {
                                this.loadData(false);
                            }
                        })
                        .catch(err => {
                            this.selection = [];
                            this.editMode = false;
                            this.dialog.loading = false;

                            err.text().then(msg => {
                                this.showErrorDialog(`${msg} (${err.status})`);
                            })
                        });
                }
            });
        },
        showDialogUpdateArchive(items) {
            // Check and filter items
            if (typeof items !== "object") return;
            if (!Array.isArray(items)) items = [items];

            items = items.filter(item => {
                var id = (typeof item.id === "number") ? item.id : 0,
                    index = (typeof item.index === "number") ? item.index : -1;

                return id > 0 && index > -1;
            });

            if (items.length === 0) return;

            // Show dialog
            var ids = items.map(item => item.id);

            this.showDialog({
                title: "Update Archive",
                content: "Update archive for selected bookmarks ? This action is irreversible.",
                mainText: "Yes",
                secondText: "No",
                mainClick: () => {
                    this.dialog.loading = true;
                    fetch("/api/archive", {
                            method: "put",
                            body: JSON.stringify(ids),
                            headers: {
                                "Content-Type": "application/json",
                            },
                        })
                        .then(response => {
                            if (!response.ok) throw response;
                            return response.json();
                        })
                        .then(json => {
                            this.selection = [];
                            this.editMode = false;
                            this.dialog.loading = false;
                            this.dialog.visible = false;

                            json.forEach(book => {
                                var item = items.find(el => el.id === book.id);
                                this.bookmarks.splice(item.index, 1, book);
                            });
                        })
                        .catch(err => {
                            this.selection = [];
                            this.editMode = false;
                            this.dialog.loading = false;

                            err.text().then(msg => {
                                this.showErrorDialog(`${msg} (${err.status})`);
                            })
                        });
                }
            });
        },
        showDialogAddTags(items) {
            // Check and filter items
            if (typeof items !== "object") return;
            if (!Array.isArray(items)) items = [items];

            items = items.filter(item => {
                var id = (typeof item.id === "number") ? item.id : 0,
                    index = (typeof item.index === "number") ? item.index : -1;

                return id > 0 && index > -1;
            });

            if (items.length === 0) return;

            // Show dialog
            this.showDialog({
                title: "Add New Tags",
                content: "Add new tags to selected bookmarks",
                fields: [{
                    name: "tags",
                    label: "Comma separated tags",
                    value: "",
                    separator: ",",
                    dictionary: this.tags.map(tag => tag.name)
                }],
                mainText: 'OK',
                secondText: 'Cancel',
                mainClick: (data) => {
                    // Validate input
                    var tags = data.tags
                        .toLowerCase()
                        .replace(/\s+/g, ' ')
                        .split(/\s*,\s*/g)
                        .filter(tag => tag.trim() !== '')
                        .map(tag => {
                            return {
                                name: tag.trim()
                            };
                        });

                    if (tags.length === 0) return;

                    // Send data
                    var request = {
                        ids: items.map(item => item.id),
                        tags: tags
                    }

                    this.dialog.loading = true;
                    fetch("/api/bookmarks/tags", {
                            method: "put",
                            body: JSON.stringify(request),
                            headers: {
                                "Content-Type": "application/json",
                            },
                        })
                        .then(response => {
                            if (!response.ok) throw response;
                            return response.json();
                        })
                        .then(json => {
                            this.selection = [];
                            this.editMode = false;
                            this.dialog.loading = false;
                            this.dialog.visible = false;

                            json.forEach(book => {
                                var item = items.find(el => el.id === book.id);
                                this.bookmarks.splice(item.index, 1, book);
                            });
                        })
                        .catch(err => {
                            this.selection = [];
                            this.editMode = false;
                            this.dialog.loading = false;

                            err.text().then(msg => {
                                this.showErrorDialog(`${msg} (${err.status})`);
                            })
                        });
                }
            });
        }
    },
    mounted() {
        var stateWatcher = (e) => {
            var state = e.state || {},
                activePage = state.activePage || "page-home",
                search = state.search || "",
                page = state.page || 1;

            if (activePage !== "page-home") return;

            this.page = page;
            this.search = search;
            this.loadData(false);
        }

        window.addEventListener('popstate', stateWatcher);
        this.$once('hook:beforeDestroy', function() {
            window.removeEventListener('popstate', stateWatcher);
        })

        this.loadData(false, true);
    }
}