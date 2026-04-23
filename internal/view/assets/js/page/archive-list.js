import { apiRequest } from "../utils/api.js";
import basePage from "./base.js";

var template = `
<div id="page-archive-list">

  <!-- Filter bar -->
  <div class="arl-filters">
    <div class="arl-filter-group" style="flex:2;min-width:180px">
      <i class="fas fa-fw fa-search"></i>
      <input type="text" placeholder="Search title, content, URL…"
        v-model.trim="filters.keyword" @keyup.enter="search">
    </div>
    <div class="arl-filter-group" style="min-width:150px">
      <i class="fas fa-fw fa-globe"></i>
      <input type="text" placeholder="Domain e.g. github.com"
        v-model.trim="filters.domain" @keyup.enter="search" list="arl-domain-list">
      <datalist id="arl-domain-list">
        <option v-for="d in knownDomains" :key="d" :value="d"></option>
      </datalist>
    </div>
    <div class="arl-filter-group" style="min-width:140px">
      <i class="fas fa-fw fa-tags"></i>
      <input type="text" placeholder="tag1, tag2…"
        v-model.trim="filters.tags" @keyup.enter="search">
    </div>
    <div class="arl-filter-group" style="min-width:130px;flex:0">
      <i class="fas fa-fw fa-calendar-alt"></i>
      <select v-model="filters.period" @change="search">
        <option value="">All time</option>
        <option value="today">Today</option>
        <option value="week">This week</option>
        <option value="month">This month</option>
        <option value="year">This year</option>
      </select>
    </div>
    <div class="arl-filter-group" style="min-width:120px;flex:0">
      <i class="fas fa-fw fa-sort-amount-down"></i>
      <select v-model="filters.sort" @change="applySort">
        <option value="newest">Newest</option>
        <option value="oldest">Oldest</option>
        <option value="title">Title A–Z</option>
      </select>
    </div>
    <button class="arl-btn arl-btn-primary" @click="search">
      <i class="fas fa-fw fa-search"></i> Search
    </button>
    <button class="arl-btn arl-btn-ghost" @click="clearFilters" title="Clear all filters">
      <i class="fas fa-fw fa-times"></i>
    </button>
  </div>

  <!-- Active filter chips -->
  <div class="arl-chips" v-if="activeChips.length">
    <div class="arl-chip" v-for="chip in activeChips" :key="chip.key">
      <i class="fas fa-fw" :class="chip.icon" style="font-size:.8em"></i>
      {{chip.label}}
      <button @click="removeChip(chip.key)" title="Remove filter">×</button>
    </div>
  </div>

  <!-- Result count -->
  <div class="arl-meta-bar" v-if="!loading">
    {{totalItems}} bookmark{{totalItems !== 1 ? 's' : ''}}
    <span v-if="maxPage > 1"> · page {{page}} of {{maxPage}}</span>
  </div>

  <!-- Spinner -->
  <div class="arl-spinner" v-if="loading">
    <i class="fas fa-fw fa-spinner fa-spin"></i>
  </div>

  <!-- Empty state -->
  <div class="arl-empty" v-else-if="!loading && bookmarks.length === 0">
    <i class="fas fa-box-open"></i>
    No bookmarks match your filters.
  </div>

  <!-- List -->
  <div class="arl-list" v-else>
    <div class="arl-item" v-for="book in sortedBookmarks" :key="book.id" @click="openBookmark(book)">

      <div class="arl-thumb">
        <img v-if="book.imageURL" :src="book.imageURL" :alt="book.title">
        <span v-else class="arl-thumb-letter">{{(book.title || '?')[0]}}</span>
      </div>

      <div class="arl-body">
        <div class="arl-title">{{book.title || book.url}}</div>
        <div class="arl-excerpt" v-if="book.excerpt">{{book.excerpt}}</div>
        <div class="arl-row-meta">
          <span class="arl-domain" @click.stop="filterByDomain(getDomain(book.url))">
            {{getDomain(book.url)}}
          </span>
          <span class="arl-tag" v-for="tag in book.tags" :key="tag.id"
            @click.stop="filterByTag(tag.name)">{{tag.name}}</span>
          <span class="arl-date">{{formatDate(book.createdAt)}}</span>
        </div>
      </div>

      <div class="arl-actions" @click.stop>
        <a v-if="book.hasArchive" :href="'bookmark/'+book.id+'/archive'" title="View Archive" target="_self">
          <i class="fas fa-archive"></i>
        </a>
        <a v-if="book.hasContent" :href="'bookmark/'+book.id+'/content'" title="View Readable" target="_self">
          <i class="fas fa-file-alt"></i>
        </a>
        <a :href="book.url" target="_blank" rel="noopener noreferrer" title="View Original">
          <i class="fas fa-external-link-alt"></i>
        </a>
      </div>

    </div>
  </div>

  <!-- Pagination -->
  <div class="arl-pagination" v-if="maxPage > 1">
    <button class="arl-page-btn" @click="goPage(page-1)" :disabled="page <= 1">
      <i class="fas fa-chevron-left"></i>
    </button>
    <button class="arl-page-btn" v-for="p in pageRange" :key="p"
      :class="{active: p === page}" @click="goPage(p)">{{p}}</button>
    <button class="arl-page-btn" @click="goPage(page+1)" :disabled="page >= maxPage">
      <i class="fas fa-chevron-right"></i>
    </button>
  </div>

</div>`;

export default {
  name: "page-archive-list",
  template: template,
  mixins: [basePage],

  data() {
    return {
      loading: false,
      bookmarks: [],
      knownDomains: [],
      page: 1,
      maxPage: 1,
      totalItems: 0,
      filters: {
        keyword: "",
        domain: "",
        tags: "",
        period: "",
        sort: "newest",
      },
    };
  },

  computed: {
    activeChips() {
      var chips = [];
      if (this.filters.keyword)
        chips.push({ key: "keyword", icon: "fa-search", label: this.filters.keyword });
      if (this.filters.domain)
        chips.push({ key: "domain", icon: "fa-globe", label: this.filters.domain });
      if (this.filters.tags)
        chips.push({ key: "tags", icon: "fa-tags", label: this.filters.tags });
      if (this.filters.period)
        chips.push({ key: "period", icon: "fa-calendar-alt", label: this.periodLabel(this.filters.period) });
      return chips;
    },

    sortedBookmarks() {
      var list = this.bookmarks.slice();

      // Client-side period filter
      if (this.filters.period) {
        var cutoff = this.periodCutoff(this.filters.period);
        list = list.filter(b => new Date(b.createdAt.replace(" ", "T") + "Z") >= cutoff);
      }

      // Sort
      if (this.filters.sort === "oldest") {
        list.sort((a, b) => a.createdAt.localeCompare(b.createdAt));
      } else if (this.filters.sort === "title") {
        list.sort((a, b) => (a.title || "").localeCompare(b.title || ""));
      } else {
        list.sort((a, b) => b.createdAt.localeCompare(a.createdAt));
      }

      return list;
    },

    pageRange() {
      var range = [], start = Math.max(1, this.page - 2), end = Math.min(this.maxPage, this.page + 2);
      for (var i = start; i <= end; i++) range.push(i);
      return range;
    },
  },

  methods: {
    async loadData() {
      this.loading = true;
      try {
        // Build keyword: merge text search + domain into one field
        var kw = [this.filters.keyword, this.filters.domain].filter(Boolean).join(" ");
        var tags = this.filters.tags.toLowerCase().replace(/\s+/g, "").split(",").filter(Boolean);

        var url = new URL("api/bookmarks", document.baseURI);
        url.search = new URLSearchParams({ keyword: kw, tags: tags.join(","), page: this.page });

        var json = await apiRequest(url);
        this.bookmarks = json.bookmarks || [];
        this.maxPage = json.maxPage || 1;
        this.page = json.page || 1;
        this.totalItems = json.maxPage
          ? (json.maxPage - 1) * 20 + this.bookmarks.length
          : this.bookmarks.length;

        // Collect domains for autocomplete
        this.bookmarks.forEach(b => {
          var d = this.getDomain(b.url);
          if (d && !this.knownDomains.includes(d)) this.knownDomains.push(d);
        });
      } catch (err) {
        this.showErrorDialog(err.message || "Failed to load bookmarks");
      } finally {
        this.loading = false;
      }
    },

    search() {
      this.page = 1;
      this.loadData();
    },

    applySort() {
      // sort is computed, no reload needed
    },

    clearFilters() {
      this.filters = { keyword: "", domain: "", tags: "", period: "", sort: "newest" };
      this.page = 1;
      this.loadData();
    },

    removeChip(key) {
      this.filters[key] = "";
      this.search();
    },

    goPage(p) {
      if (p < 1 || p > this.maxPage) return;
      this.page = p;
      this.loadData();
      window.scrollTo({ top: 0, behavior: "smooth" });
    },

    openBookmark(book) {
      if (book.hasArchive) {
        window.location.href = new URL("bookmark/" + book.id + "/archive", document.baseURI);
      } else if (book.hasContent) {
        window.location.href = new URL("bookmark/" + book.id + "/content", document.baseURI);
      } else {
        window.open(book.url, "_blank", "noopener,noreferrer");
      }
    },

    filterByDomain(domain) {
      this.filters.domain = domain;
      this.search();
    },

    filterByTag(name) {
      var existing = this.filters.tags.split(",").map(t => t.trim()).filter(Boolean);
      if (!existing.includes(name)) existing.push(name);
      this.filters.tags = existing.join(", ");
      this.search();
    },

    getDomain(url) {
      try { return new URL(url).hostname.replace(/^www\./, ""); }
      catch (_) { return url; }
    },

    formatDate(str) {
      if (!str) return "";
      try {
        var d = new Date(str.replace(" ", "T") + "Z");
        return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
      } catch (_) { return str; }
    },

    periodLabel(p) {
      return { today: "Today", week: "This week", month: "This month", year: "This year" }[p] || p;
    },

    periodCutoff(p) {
      var now = new Date();
      if (p === "today")  { var d = new Date(now); d.setHours(0,0,0,0); return d; }
      if (p === "week")   { var d = new Date(now); d.setDate(d.getDate() - 7); return d; }
      if (p === "month")  { var d = new Date(now); d.setMonth(d.getMonth() - 1); return d; }
      if (p === "year")   { var d = new Date(now); d.setFullYear(d.getFullYear() - 1); return d; }
      return new Date(0);
    },
  },

  mounted() {
    this.loadData();
  },
};
