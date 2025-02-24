var template = `
<div id="page-home">
    <div class="page-header">
        <input type="text" placeholder="Search url, keyword or tags" v-model.trim="search" @focus="$event.target.select()" @keyup.enter="searchBookmarks"/>
        <a title="Refresh storage" @click="reloadData">
            <i class="fas fa-fw fa-sync-alt" :class="loading && 'fa-spin'"></i>
        </a>
        <a v-if="activeAccount.owner" title="Add new bookmark" @click="showDialogAdd">
            <i class="fas fa-fw fa-plus-circle"></i>
        </a>
        <a v-if="tags.length > 0" title="Show tags" @click="showDialogTags">
            <i class="fas fa-fw fa-tags"></i>
        </a>
        <a v-if="activeAccount.owner" title="Batch edit" @click="toggleEditMode">
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
        <a title="Update archives" @click="showDialogUpdateCache(selection)">
            <i class="fas fa-fw fa-cloud-download-alt"></i>
        </a>
        <a title="Download ebooks" @click="ebookGenerate(selection)">
            <i class="fas fa-fw fa-book"></i>
        </a>
        <a title="Cancel" @click="toggleEditMode">
            <i class="fas fa-fw fa-times"></i>
        </a>
    </div>
    <p class="empty-message" v-if="!loading && listIsEmpty">No saved bookmarks yet :(</p>
    <div id="bookmarks-grid" ref="bookmarksGrid" :class="{list: appOptions.ListMode}">
        <pagination-box v-if="maxPage > 1"
            :page="page"
            :maxPage="maxPage"
            :editMode="editMode"
            @change="changePage">
        </pagination-box>
        <bookmark-item v-for="(book, index) in bookmarks"
            :id="book.id"
            :url="book.url"
            :title="book.title"
            :excerpt="book.excerpt"
            :public="book.public"
            :imageURL="book.imageURL"
            :modifiedAt="book.modifiedAt"
            :hasContent="book.hasContent"
            :hasArchive="book.hasArchive"
            :hasEbook="book.hasEbook"
            :tags="book.tags"
            :index="index"
            :key="book.id"
            :editMode="editMode"
            :ShowId="appOptions.ShowId"
            :ListMode="appOptions.ListMode"
            :HideThumbnail="appOptions.HideThumbnail"
            :HideExcerpt="appOptions.HideExcerpt"
            :selected="isSelected(book.id)"
            :menuVisible="activeAccount.owner"
            @select="toggleSelection"
            @tag-clicked="bookmarkTagClicked"
            @edit="showDialogEdit"
            @delete="showDialogDelete"
            @generate-ebook="ebookGenerate"
            @update="showDialogUpdateCache">
        </bookmark-item>
        <pagination-box v-if="maxPage > 1"
            :page="page"
            :maxPage="maxPage"
            :editMode="editMode"
            @change="changePage">
        </pagination-box>
    </div>
    <div class="loading-overlay" v-if="loading"><i class="fas fa-fw fa-spin fa-spinner"></i></div>
    <custom-dialog id="dialog-tags" v-bind="dialogTags">
        <a @click="filterTag('*')">(all tagged)</a>
        <a @click="filterTag('*', true)">(all untagged)</a>
        <a v-for="tag in tags" @click="dialogTagClicked($event, tag)">
            #{{tag.name}}<span>{{tag.nBookmarks}}</span>
        </a>
    </custom-dialog>
    <custom-dialog v-bind="dialog"/>
</div>`;

import paginationBox from "../component/pagination.js";
import bookmarkItem from "../component/bookmark.js";
import customDialog from "../component/dialog.js";
import basePage from "./base.js";
import EventBus from "../component/eventBus.js";

Vue.prototype.$bus = EventBus;

export default {
	template: template,
	mixins: [basePage],
	components: {
		bookmarkItem,
		paginationBox,
		customDialog,
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

			dialogTags: {
				visible: false,
				editMode: false,
				title: "Existing Tags",
				mainText: "OK",
				secondText: "Rename Tags",
				mainClick: () => {
					if (this.dialogTags.editMode) {
						this.dialogTags.editMode = false;
					} else {
						this.dialogTags.visible = false;
					}
				},
				secondClick: () => {
					this.dialogTags.editMode = true;
				},
				escPressed: () => {
					this.dialogTags.visible = false;
					this.dialogTags.editMode = false;
				},
			},
		};
	},
	computed: {
		listIsEmpty() {
			return this.bookmarks.length <= 0;
		},
	},
	watch: {
		"dialogTags.editMode"(editMode) {
			if (editMode) {
				this.dialogTags.title = "Rename Tags";
				this.dialogTags.mainText = "Cancel";
				this.dialogTags.secondText = "";
			} else {
				this.dialogTags.title = "Existing Tags";
				this.dialogTags.mainText = "OK";
				this.dialogTags.secondText = "Rename Tags";
			}
		},
	},
	methods: {
		clearHomePage() {
			this.search = "";
			this.searchBookmarks();
		},
		reloadData() {
			if (this.loading) return;
			this.page = 1;
			this.search = "";
			this.loadData(true, true);
		},
		loadData(saveState, fetchTags) {
			if (this.loading) return;

			// Set default args
			saveState = typeof saveState === "boolean" ? saveState : true;
			fetchTags = typeof fetchTags === "boolean" ? fetchTags : false;

			// Parse search query
			var keyword = this.search,
				rxExcludeTagA = /(^|\s)-tag:["']([^"']+)["']/i, // -tag:"with space"
				rxExcludeTagB = /(^|\s)-tag:(\S+)/i, // -tag:without-space
				rxIncludeTagA = /(^|\s)tag:["']([^"']+)["']/i, // tag:"with space"
				rxIncludeTagB = /(^|\s)tag:(\S+)/i, // tag:without-space
				tags = [],
				excludedTags = [],
				rxResult;

			// Get excluded tag first, while also removing it from keyword
			while ((rxResult = rxExcludeTagA.exec(keyword))) {
				keyword = keyword.replace(rxResult[0], "");
				excludedTags.push(rxResult[2]);
			}

			while ((rxResult = rxExcludeTagB.exec(keyword))) {
				keyword = keyword.replace(rxResult[0], "");
				excludedTags.push(rxResult[2]);
			}

			// Get included tags
			while ((rxResult = rxIncludeTagA.exec(keyword))) {
				keyword = keyword.replace(rxResult[0], "");
				tags.push(rxResult[2]);
			}

			while ((rxResult = rxIncludeTagB.exec(keyword))) {
				keyword = keyword.replace(rxResult[0], "");
				tags.push(rxResult[2]);
			}

			// Trim keyword
			keyword = keyword.trim().replace(/\s+/g, " ");

			// Prepare URL for API
			var url = new URL("api/bookmarks", document.baseURI);
			url.search = new URLSearchParams({
				keyword: keyword,
				tags: tags.join(","),
				exclude: excludedTags.join(","),
				page: this.page,
			});

			// Fetch data from API
			var skipFetchTags = Error("skip fetching tags");

			this.loading = true;
			fetch(url, {
				headers: {
					"Content-Type": "application/json",
					Authorization: "Bearer " + localStorage.getItem("shiori-token"),
				},
			})
				.then((response) => {
					if (!response.ok) throw response;
					return response.json();
				})
				.then((json) => {
					// Set data
					this.page = json.page;
					this.maxPage = json.maxPage;
					this.bookmarks = json.bookmarks;

					// Save state and change URL if needed
					if (saveState) {
						var history = {
							activePage: "page-home",
							search: this.search,
							page: this.page,
						};

						var url = new Url(document.baseURI);
						url.hash = "home";
						url.clearQuery();
						if (this.page > 1) url.query.page = this.page;
						if (this.search !== "") url.query.search = this.search;

						window.history.pushState(history, "page-home", url);
					}

					// Fetch tags if requested
					if (fetchTags) {
						return fetch(new URL("api/v1/tags", document.baseURI), {
							headers: {
								"Content-Type": "application/json",
								Authorization: "Bearer " + localStorage.getItem("shiori-token"),
							},
						});
					} else {
						this.loading = false;
						throw skipFetchTags;
					}
				})
				.then((response) => {
					if (!response.ok) throw response;
					return response.json();
				})
				.then((json) => {
					this.tags = json.message;
					this.loading = false;
				})
				.catch((err) => {
					this.loading = false;

					if (err !== skipFetchTags) {
						this.getErrorMessage(err).then((msg) => {
							this.showErrorDialog(msg);
						});
					}
				});
		},
		searchBookmarks() {
			this.page = 1;
			this.loadData();
		},
		changePage(page) {
			this.page = page;
			this.$refs.bookmarksGrid.scrollTop = 0;
			this.loadData();
		},
		toggleEditMode() {
			this.selection = [];
			this.editMode = !this.editMode;
		},
		toggleSelection(item) {
			var idx = this.selection.findIndex((el) => el.id === item.id);
			if (idx === -1) this.selection.push(item);
			else this.selection.splice(idx, 1);
		},
		isSelected(bookId) {
			return this.selection.findIndex((el) => el.id === bookId) > -1;
		},
		dialogTagClicked(event, tag) {
			if (!this.dialogTags.editMode) {
				this.filterTag(tag.name, event.altKey);
			} else {
				this.dialogTags.visible = false;
				this.showDialogRenameTag(tag);
			}
		},
		bookmarkTagClicked(event, tagName) {
			this.filterTag(tagName, event.altKey);
		},
		filterTag(tagName, excludeMode) {
			// Set default parameter
			excludeMode = typeof excludeMode === "boolean" ? excludeMode : false;

			if (this.dialogTags.editMode) {
				return;
			}

			if (tagName === "*") {
				this.search = excludeMode ? "-tag:*" : "tag:*";
				this.page = 1;
				this.loadData();
				return;
			}

			var rxSpace = /\s+/g,
				includeTag = rxSpace.test(tagName)
					? `tag:"${tagName}"`
					: `tag:${tagName}`,
				excludeTag = "-" + includeTag,
				rxIncludeTag = new RegExp(`(^|\\s)${includeTag}`, "ig"),
				rxExcludeTag = new RegExp(`(^|\\s)${excludeTag}`, "ig"),
				search = this.search;

			search = search.replace("-tag:*", "");
			search = search.replace("tag:*", "");
			search = search.trim();

			if (excludeMode) {
				if (rxExcludeTag.test(search)) {
					return;
				}

				if (rxIncludeTag.test(search)) {
					this.search = search.replace(rxIncludeTag, "$1" + excludeTag);
				} else {
					search += ` ${excludeTag}`;
					this.search = search.trim();
				}
			} else {
				if (rxIncludeTag.test(search)) {
					return;
				}

				if (rxExcludeTag.test(search)) {
					this.search = search.replace(rxExcludeTag, "$1" + includeTag);
				} else {
					search += ` ${includeTag}`;
					this.search = search.trim();
				}
			}

			this.page = 1;
			this.loadData();
		},
		showDialogAdd(values) {
			if (values === undefined) {
				values = {};
			}

			this.showDialog({
				title: "New Bookmark",
				content: "Create a new bookmark",
				fields: [
					{
						name: "url",
						label: "Url, start with http://...",
						value: values.url || "",
					},
					{
						name: "title",
						label: "Custom title (optional)",
						value: values.title || "",
					},
					{
						name: "excerpt",
						label: "Custom excerpt (optional)",
						type: "area",
						value: values.excerpt || "",
					},
					{
						name: "tags",
						label: "Comma separated tags (optional)",
						separator: ",",
						dictionary: this.tags.map((tag) => tag.name),
					},
					{
						name: "create_archive",
						label: "Create archive",
						type: "check",
						value: this.appOptions.UseArchive,
					},
					{
						name: "create_ebook",
						label: "Create Ebook",
						type: "check",
						value: this.appOptions.CreateEbook,
					},
					{
						name: "makePublic",
						label: "Make bookmark publicly available",
						type: "check",
						value: this.appOptions.MakePublic,
					},
				],
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
						.filter((tag) => tag.trim() !== "")
						.map((tag) => {
							return {
								name: tag.trim(),
							};
						});

					// Send data
					var data = {
						url: data.url.trim(),
						title: data.title.trim(),
						excerpt: data.excerpt.trim(),
						public: data.makePublic ? 1 : 0,
						tags: tags,
						create_archive: data.create_archive,
						create_ebook: data.create_ebook,
					};

					this.dialog.loading = true;
					fetch(new URL("api/bookmarks", document.baseURI), {
						method: "post",
						body: JSON.stringify(data),
						headers: {
							"Content-Type": "application/json",
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response.json();
						})
						.then((json) => {
							this.dialog.loading = false;
							this.dialog.visible = false;
							this.bookmarks.splice(0, 0, json);
						})
						.catch((err) => {
							this.dialog.loading = false;
							this.getErrorMessage(err).then((msg) => {
								this.showErrorDialog(msg);
							});
						});
				},
			});
		},
		showDialogEdit(item) {
			// Check the item
			if (typeof item !== "object") return;

			var id = typeof item.id === "number" ? item.id : 0,
				index = typeof item.index === "number" ? item.index : -1;

			if (id < 1 || index < 0) return;

			// Get the existing bookmark value
			var book = JSON.parse(JSON.stringify(this.bookmarks[index])),
				strTags = book.tags.map((tag) => tag.name).join(", ");

			this.showDialog({
				title: "Edit Bookmark",
				content: "Edit the bookmark's data",
				showLabel: true,
				fields: [
					{
						name: "url",
						label: "Url",
						value: book.url,
					},
					{
						name: "title",
						label: "Title",
						value: book.title,
					},
					{
						name: "excerpt",
						label: "Excerpt",
						type: "area",
						value: book.excerpt,
					},
					{
						name: "tags",
						label: "Tags",
						value: strTags,
						separator: ",",
						dictionary: this.tags.map((tag) => tag.name),
					},
					{
						name: "makePublic",
						label: "Make bookmark publicly available",
						type: "check",
						value: book.public >= 1,
					},
				],
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
						.filter((tag) => tag.trim() !== "")
						.map((tag) => {
							return {
								name: tag.trim(),
							};
						});

					// Set new data
					book.url = data.url.trim();
					book.title = data.title.trim();
					book.excerpt = data.excerpt.trim();
					book.public = data.makePublic ? 1 : 0;
					book.tags = tags;

					// Send data
					this.dialog.loading = true;
					fetch(new URL("api/bookmarks", document.baseURI), {
						method: "put",
						body: JSON.stringify(book),
						headers: {
							"Content-Type": "application/json",
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response.json();
						})
						.then((json) => {
							this.dialog.loading = false;
							this.dialog.visible = false;
							this.bookmarks.splice(index, 1, json);
						})
						.catch((err) => {
							this.dialog.loading = false;
							this.getErrorMessage(err).then((msg) => {
								this.showErrorDialog(msg);
							});
						});
				},
			});
		},
		showDialogDelete(items) {
			// Check and filter items
			if (typeof items !== "object") return;
			if (!Array.isArray(items)) items = [items];

			items = items.filter((item) => {
				var id = typeof item.id === "number" ? item.id : 0,
					index = typeof item.index === "number" ? item.index : -1;

				return id > 0 && index > -1;
			});

			if (items.length === 0) return;

			// Split ids and indices
			var ids = items.map((item) => item.id),
				indices = items.map((item) => item.index).sort((a, b) => b - a);

			// Create title and content
			var title = "Delete Bookmarks",
				content =
					"Delete the selected bookmarks ? This action is irreversible.";

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
					fetch(new URL("api/bookmarks", document.baseURI), {
						method: "delete",
						body: JSON.stringify(ids),
						headers: {
							"Content-Type": "application/json",
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response;
						})
						.then(() => {
							this.selection = [];
							this.editMode = false;
							this.dialog.loading = false;
							this.dialog.visible = false;
							indices.forEach((index) => this.bookmarks.splice(index, 1));

							if (this.bookmarks.length < 20) {
								this.loadData(false);
							}
						})
						.catch((err) => {
							this.selection = [];
							this.editMode = false;
							this.dialog.loading = false;

							this.getErrorMessage(err).then((msg) => {
								this.showErrorDialog(msg);
							});
						});
				},
			});
		},
		ebookGenerate(items) {
			// Check and filter items
			if (typeof items !== "object") return;
			if (!Array.isArray(items)) items = [items];

			items = items.filter((item) => {
				var id = typeof item.id === "number" ? item.id : 0,
					index = typeof item.index === "number" ? item.index : -1;

				return id > 0 && index > -1;
			});

			if (items.length === 0) return;

			// define variable and send request
			var ids = items.map((item) => item.id);
			var data = {
				ids: ids,
				create_archive: false,
				keep_metadata: true,
				create_ebook: true,
				skip_exist: true,
			};
			this.loading = true;
			fetch(new URL("api/v1/bookmarks/cache", document.baseURI), {
				method: "put",
				body: JSON.stringify(data),
				headers: {
					"Content-Type": "application/json",
					Authorization: "Bearer " + localStorage.getItem("shiori-token"),
				},
			})
				.then((response) => {
					if (!response.ok) throw response;
					return response.json();
				})
				.then((json) => {
					this.selection = [];
					this.editMode = false;
					json.message.forEach((book) => {
						// download ebooks
						const id = book.id;
						if (book.hasEbook) {
							const ebook_url = new URL(
								`bookmark/${id}/ebook`,
								document.baseURI,
							);
							const downloadLink = document.createElement("a");
							downloadLink.href = ebook_url.toString();
							downloadLink.download = `${book.title}.epub`;
							downloadLink.click();
						}

						var item = items.find((el) => el.id === book.id);
						this.bookmarks.splice(item.index, 1, book);
					});
				})
				.catch((err) => {
					this.selection = [];
					this.editMode = false;
					this.getErrorMessage(err).then((msg) => {
						this.showErrorDialog(msg);
					});
				})
				.finally(() => {
					this.loading = false;
				});
		},
		showDialogUpdateCache(items) {
			// Check and filter items
			if (typeof items !== "object") return;
			if (!Array.isArray(items)) items = [items];

			items = items.filter((item) => {
				var id = typeof item.id === "number" ? item.id : 0,
					index = typeof item.index === "number" ? item.index : -1;

				return id > 0 && index > -1;
			});

			if (items.length === 0) return;

			// Show dialog
			var ids = items.map((item) => item.id);

			this.showDialog({
				title: "Update Cache",
				content:
					"Update cache for selected bookmarks ? This action is irreversible.",
				fields: [
					{
						name: "keep_metadata",
						label: "Keep the old title and excerpt",
						type: "check",
						value: this.appOptions.KeepMetadata,
					},
					{
						name: "create_archive",
						label: "Update archive as well",
						type: "check",
						value: this.appOptions.UseArchive,
					},
					{
						name: "create_ebook",
						label: "Update Ebook as well",
						type: "check",
						value: this.appOptions.CreateEbook,
					},
				],
				mainText: "Yes",
				secondText: "No",
				mainClick: (data) => {
					var data = {
						ids: ids,
						create_archive: data.create_archive,
						keep_metadata: data.keep_metadata,
						create_ebook: data.create_ebook,
						skip_exist: false,
					};

					this.dialog.loading = true;
					fetch(new URL("api/v1/bookmarks/cache", document.baseURI), {
						method: "put",
						body: JSON.stringify(data),
						headers: {
							"Content-Type": "application/json",
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response.json();
						})
						.then((json) => {
							this.selection = [];
							this.editMode = false;
							this.dialog.loading = false;
							this.dialog.visible = false;

							let faildedUpdateArchives = [];
							let faildedCreateEbook = [];
							json.message.forEach((book) => {
								var item = items.find((el) => el.id === book.id);
								this.bookmarks.splice(item.index, 1, book);

								if (data.create_archive && !book.hasArchive) {
									faildedUpdateArchives.push(book.id);
									console.error(
										"can't update archive for bookmark id",
										book.id,
									);
								}
								if (data.create_ebook && !book.hasEbook) {
									faildedCreateEbook.push(book.id);
									console.error("can't update ebook for bookmark id:", book.id);
								}
							});
							if (
								faildedCreateEbook.length > 0 ||
								faildedUpdateArchives.length > 0
							) {
								this.showDialog({
									title: `Bookmarks Id that Update Action Faild`,
									content: `Not all bookmarks could have their contents updated, but no files were overwritten.`,
									mainText: "OK",
									mainClick: () => {
										this.dialog.visible = false;
									},
								});
							}
						})
						.catch((err) => {
							this.selection = [];
							this.editMode = false;
							this.dialog.loading = false;

							this.getErrorMessage(err).then((msg) => {
								this.showErrorDialog(msg);
							});
						});
				},
			});
		},
		showDialogAddTags(items) {
			// Check and filter items
			if (typeof items !== "object") return;
			if (!Array.isArray(items)) items = [items];

			items = items.filter((item) => {
				var id = typeof item.id === "number" ? item.id : 0,
					index = typeof item.index === "number" ? item.index : -1;

				return id > 0 && index > -1;
			});

			if (items.length === 0) return;

			// Show dialog
			this.showDialog({
				title: "Add New Tags",
				content: "Add new tags to selected bookmarks",
				fields: [
					{
						name: "tags",
						label: "Comma separated tags",
						value: "",
						separator: ",",
						dictionary: this.tags.map((tag) => tag.name),
					},
				],
				mainText: "OK",
				secondText: "Cancel",
				mainClick: (data) => {
					// Validate input
					var tags = data.tags
						.toLowerCase()
						.replace(/\s+/g, " ")
						.split(/\s*,\s*/g)
						.filter((tag) => tag.trim() !== "")
						.map((tag) => {
							return {
								name: tag.trim(),
							};
						});

					if (tags.length === 0) return;

					// Send data
					var request = {
						ids: items.map((item) => item.id),
						tags: tags,
					};

					this.dialog.loading = true;
					fetch(new URL("api/bookmarks/tags", document.baseURI), {
						method: "put",
						body: JSON.stringify(request),
						headers: {
							"Content-Type": "application/json",
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response.json();
						})
						.then((json) => {
							this.selection = [];
							this.editMode = false;
							this.dialog.loading = false;
							this.dialog.visible = false;

							json.forEach((book) => {
								var item = items.find((el) => el.id === book.id);
								this.bookmarks.splice(item.index, 1, book);
							});
						})
						.catch((err) => {
							this.selection = [];
							this.editMode = false;
							this.dialog.loading = false;

							this.getErrorMessage(err).then((msg) => {
								this.showErrorDialog(msg);
							});
						});
				},
			});
		},
		showDialogTags() {
			this.dialogTags.visible = true;
			this.dialogTags.editMode = false;
			this.dialogTags.secondText = this.activeAccount.owner
				? "Rename Tags"
				: "";
		},
		showDialogRenameTag(tag) {
			this.showDialog({
				title: "Rename Tag",
				content: `Change the name for tag "#${tag.name}"`,
				fields: [
					{
						name: "newName",
						label: "New tag name",
						value: tag.name,
					},
				],
				mainText: "OK",
				secondText: "Cancel",
				secondClick: () => {
					this.dialog.visible = false;
					this.dialogTags.visible = true;
				},
				escPressed: () => {
					this.dialog.visible = false;
					this.dialogTags.visible = true;
				},
				mainClick: (data) => {
					// Save the old query
					var rxSpace = /\s+/g,
						oldTagQuery = rxSpace.test(tag.name)
							? `"#${tag.name}"`
							: `#${tag.name}`,
						newTagQuery = rxSpace.test(data.newName)
							? `"#${data.newName}"`
							: `#${data.newName}`;

					// Send data
					var newData = {
						id: tag.id,
						name: data.newName,
					};

					this.dialog.loading = true;
					fetch(new URL("api/tags", document.baseURI), {
						method: "PUT",
						body: JSON.stringify(newData),
						headers: {
							"Content-Type": "application/json",
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response.json();
						})
						.then(() => {
							tag.name = data.newName;

							this.dialog.loading = false;
							this.dialog.visible = false;
							this.dialogTags.visible = true;
							this.dialogTags.editMode = false;
							this.tags.sort((a, b) => {
								var aName = a.name.toLowerCase(),
									bName = b.name.toLowerCase();

								if (aName < bName) return -1;
								else if (aName > bName) return 1;
								else return 0;
							});

							if (this.search.includes(oldTagQuery)) {
								this.search = this.search.replace(oldTagQuery, newTagQuery);
								this.loadData();
							}
						})
						.catch((err) => {
							this.dialog.loading = false;
							this.dialogTags.visible = false;
							this.dialogTags.editMode = false;
							this.getErrorMessage(err).then((msg) => {
								this.showErrorDialog(msg);
							});
						});
				},
			});
		},
	},
	mounted() {
		this.$bus.$on("clearHomePage", () => {
			this.clearHomePage();
		});
		// Prepare history state watcher
		var stateWatcher = (e) => {
			var state = e.state || {},
				activePage = state.activePage || "page-home",
				search = state.search || "",
				page = state.page || 1;

			if (activePage !== "page-home") return;

			this.page = page;
			this.search = search;
			this.loadData(false);
		};

		window.addEventListener("popstate", stateWatcher);
		this.$once("hook:beforeDestroy", () => {
			window.removeEventListener("popstate", stateWatcher);
		});

		// Set initial parameter
		var url = new Url();
		this.search = url.query.search || "";
		this.page = url.query.page || 1;

		var isSharing =
			url.query.url !== undefined || url.query.excerpt !== undefined;
		if (isSharing) {
			// this is what the spec says
			var shareData = {
				url: url.query.url,
				excerpt: url.query.excerpt,
				title: url.query.title,
			};

			// In my testing sharing from chrome and ff focus, this is how data arrives
			if (shareData.url === undefined) {
				shareData.url = url.query.excerpt;
				shareData.title = url.query.title;
				shareData.excerpt = "";
			}

			this.showDialogAdd(shareData);
			var history = {
				activePage: "page-home",
				search: this.search,
				page: this.page,
			};

			var url = new Url(document.baseURI);
			url.hash = "home";
			url.clearQuery();
			window.history.replaceState(history, "page-home", url);
		}

		this.loadData(false, true);
	},
};
