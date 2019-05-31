var template = `
<div id="page-setting">
    <div class="page-header">
        <p>Settings</p>
        <a href="#" title="Refresh setting">
            <i class="fas fa-fw fa-sync-alt"></i>
        </a>
    </div>
    <div class="setting-container">
        <details open class="setting-group" id="setting-display">
            <summary>Display</summary>
            <label>
                <input type="checkbox" v-model="displayOptions.showId" @change="saveSetting">
                Show bookmark's ID
            </label>
            <label>
                <input type="checkbox" v-model="displayOptions.listMode" @change="saveSetting">
                Display bookmarks as list
            </label>
            <label>
                <input type="checkbox" v-model="displayOptions.nightMode" @change="saveSetting">
                Use dark theme
            </label>
        </details>
    </div>
    <div class="loading-overlay" v-if="loading"><i class="fas fa-fw fa-spin fa-spinner"></i></div>
    <custom-dialog v-bind="dialog"/>
</div>`;

import customDialog from "../component/dialog.js";
import basePage from "./base.js";

export default {
    template: template,
    mixins: [basePage],
    components: {
        customDialog
    },
    data() {
        return {
            loading: false,
        }
    },
    methods: {
        saveSetting() {
            this.$emit("setting-changed", {
                showId: this.displayOptions.showId,
                listMode: this.displayOptions.listMode,
                nightMode: this.displayOptions.nightMode,
            });
        }
    }
}