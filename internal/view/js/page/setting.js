var template = `
<div id="page-setting">
    <h1 class="page-header">Settings</h1>
    <div class="setting-container">
        <details open class="setting-group" id="setting-display">
            <summary>显示</summary>
            <label>
                <input type="checkbox" v-model="appOptions.showId" @change="saveSetting">
                显示书签ID
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.listMode" @change="saveSetting">
                将书签显示为列表
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.hideThumbnail" @change="saveSetting">
                隐藏缩略图
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.hideExcerpt" @change="saveSetting">
                隐藏书签的节选
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.nightMode" @change="saveSetting">
                使用深色主题
            </label>
        </details>
        <details v-if="activeAccount.owner" open class="setting-group" id="setting-bookmarks">
            <summary>书签</summary>
            <label>
                <input type="checkbox" v-model="appOptions.keepMetadata" @change="saveSetting">
                更新时保持书签的元数据
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.useArchive" @change="saveSetting">
                创建默认存档
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.makePublic" @change="saveSetting">
                默认情况下设置存档为公开
            </label>
        </details>
        <details v-if="activeAccount.owner" open class="setting-group" id="setting-accounts">
            <summary>账户</summary>
            <ul>
                <li v-if="accounts.length === 0">No accounts registered</li>
                <li v-for="(account, idx) in accounts">
                    <p>{{account.username}}
                        <span v-if="account.owner" class="account-level">(owner)</span>
                    </p>
                    <a title="更改密码" @click="showDialogChangePassword(account)">
                        <i class="fa fas fa-fw fa-key"></i>
                    </a>
                    <a title="删除用户" @click="showDialogDeleteAccount(account, idx)">
                        <i class="fa fas fa-fw fa-trash-alt"></i>
                    </a>
                </li>
            </ul>
            <div class="setting-group-footer">
                <a @click="loadAccounts">刷新账户列表</a>
                <a v-if="activeAccount.owner" @click="showDialogNewAccount">添加新用户</a>
            </div>
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
			accounts: []
		}
	},
	methods: {
		saveSetting() {
			this.$emit("setting-changed", {
				showId: this.appOptions.showId,
				listMode: this.appOptions.listMode,
				hideThumbnail: this.appOptions.hideThumbnail,
				hideExcerpt: this.appOptions.hideExcerpt,
				nightMode: this.appOptions.nightMode,
				keepMetadata: this.appOptions.keepMetadata,
				useArchive: this.appOptions.useArchive,
				makePublic: this.appOptions.makePublic,
			});
		},
		loadAccounts() {
			if (this.loading) return;

			this.loading = true;
			fetch(new URL("api/accounts", document.baseURI))
				.then(response => {
					if (!response.ok) throw response;
					return response.json();
				})
				.then(json => {
					this.loading = false;
					this.accounts = json;
				})
				.catch(err => {
					this.loading = false;
					this.getErrorMessage(err).then(msg => {
						this.showErrorDialog(msg);
					})
				});
		},
		showDialogNewAccount() {
			this.showDialog({
				title: "New Account",
				content: "Input new account's data :",
				fields: [{
					name: "username",
					label: "Username",
					value: "",
				}, {
					name: "password",
					label: "Password",
					type: "password",
					value: "",
				}, {
					name: "repeat",
					label: "Repeat password",
					type: "password",
					value: "",
				}, {
					name: "visitor",
					label: "This account is for visitor",
					type: "check",
					value: false,
				}],
				mainText: "OK",
				secondText: "Cancel",
				mainClick: (data) => {
					if (data.username === "") {
						this.showErrorDialog("Username must not empty");
						return;
					}

					if (data.password === "") {
						this.showErrorDialog("Password must not empty");
						return;
					}

					if (data.password !== data.repeat) {
						this.showErrorDialog("Password does not match");
						return;
					}

					var request = {
						username: data.username,
						password: data.password,
						owner: !data.visitor,
					}

					this.dialog.loading = true;
					fetch(new URL("api/accounts", document.baseURI), {
						method: "post",
						body: JSON.stringify(request),
						headers: {
							"Content-Type": "application/json",
						}
					}).then(response => {
						if (!response.ok) throw response;
						return response;
					}).then(() => {
						this.dialog.loading = false;
						this.dialog.visible = false;

						this.accounts.push({ username: data.username, owner: !data.visitor });
						this.accounts.sort((a, b) => {
							var nameA = a.username.toLowerCase(),
								nameB = b.username.toLowerCase();

							if (nameA < nameB) {
								return -1;
							}

							if (nameA > nameB) {
								return 1;
							}

							return 0;
						});
					}).catch(err => {
						this.dialog.loading = false;
						this.getErrorMessage(err).then(msg => {
							this.showErrorDialog(msg);
						})
					});
				}
			});
		},
		showDialogChangePassword(account) {
			this.showDialog({
				title: "Change Password",
				content: "Input new password :",
				fields: [{
					name: "oldPassword",
					label: "Old password",
					type: "password",
					value: "",
				}, {
					name: "password",
					label: "New password",
					type: "password",
					value: "",
				}, {
					name: "repeat",
					label: "Repeat password",
					type: "password",
					value: "",
				}],
				mainText: "OK",
				secondText: "Cancel",
				mainClick: (data) => {
					if (data.oldPassword === "") {
						this.showErrorDialog("Old password must not empty");
						return;
					}

					if (data.password === "") {
						this.showErrorDialog("New password must not empty");
						return;
					}

					if (data.password !== data.repeat) {
						this.showErrorDialog("Password does not match");
						return;
					}

					var request = {
						username: account.username,
						oldPassword: data.oldPassword,
						newPassword: data.password,
						owner: account.owner,
					}

					this.dialog.loading = true;
					fetch(new URL("api/accounts", document.baseURI), {
						method: "put",
						body: JSON.stringify(request),
						headers: {
							"Content-Type": "application/json",
						},
					}).then(response => {
						if (!response.ok) throw response;
						return response;
					}).then(() => {
						this.dialog.loading = false;
						this.dialog.visible = false;
					}).catch(err => {
						this.dialog.loading = false;
						this.getErrorMessage(err).then(msg => {
							this.showErrorDialog(msg);
						})
					});
				}
			});
		},
		showDialogDeleteAccount(account, idx) {
			this.showDialog({
				title: "Delete Account",
				content: `Delete account "${account.username}" ?`,
				mainText: "Yes",
				secondText: "No",
				mainClick: () => {
					this.dialog.loading = true;
					fetch(`/api/accounts`, {
						method: "delete",
						body: JSON.stringify([account.username]),
						headers: {
							"Content-Type": "application/json",
						},
					}).then(response => {
						if (!response.ok) throw response;
						return response;
					}).then(() => {
						this.dialog.loading = false;
						this.dialog.visible = false;
						this.accounts.splice(idx, 1);
					}).catch(err => {
						this.dialog.loading = false;
						this.getErrorMessage(err).then(msg => {
							this.showErrorDialog(msg);
						})
					});
				}
			});
		},
	},
	mounted() {
		this.loadAccounts();
	}
}
