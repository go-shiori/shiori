var template = `
<div class="pagination-box">
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
</div>`

export default {
	template: template,
	props: {
		page: Number,
		maxPage: Number,
		editMode: Boolean,
	},
	methods: {
		changePage(page) {
			page = parseInt(page, 10) || 0;
			if (page >= this.maxPage) page = this.maxPage;
			else if (page <= 1) page = 1;

			this.$emit("change", page);
		}
	}
}