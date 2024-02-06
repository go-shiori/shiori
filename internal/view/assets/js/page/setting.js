var template = `
<div id="page-setting">
    <h1 class="page-header">Settings</h1>
    <div class="setting-container">
        <details open class="setting-group" id="setting-display">
            <summary>Display</summary>
            <label>
                <input type="checkbox" v-model="appOptions.ShowId" @change="saveSetting">
                Show bookmark's ID
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.ListMode" @change="saveSetting">
                Display bookmarks as list
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.HideThumbnail" @change="saveSetting">
                Hide thumbnail image
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.HideExcerpt" @change="saveSetting">
                Hide bookmark's excerpt
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.NightMode" @change="saveSetting">
                Use dark theme
            </label>
        </details>
        <details v-if="activeAccount.owner" open class="setting-group" id="setting-bookmarks">
            <summary>Bookmarks</summary>
            <label>
                <input type="checkbox" v-model="appOptions.KeepMetadata" @change="saveSetting">
                Keep bookmark's metadata when updating
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.UseArchive" @change="saveSetting">
                Create archive by default
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.CreateEbook" @change="saveSetting">
                Create ebook by default
            </label>
            <label>
                <input type="checkbox" v-model="appOptions.MakePublic" @change="saveSetting">
                Make archive publicly available by default
            </label>
        </details>
        <details v-if="activeAccount.owner" open class="setting-group" id="setting-accounts">
            <summary>Accounts</summary>
            <ul>
                <li v-if="accounts.length === 0">No accounts registered</li>
                <li v-for="(account, idx) in accounts">
                    <p>{{account.username}}
                        <span v-if="account.owner" class="account-level">(owner)</span>
                    </p>
                    <a title="Change password" @click="showDialogChangePassword(account)">
                        <i class="fa fas fa-fw fa-key"></i>
                    </a>
                    <a title="Delete account" @click="showDialogDeleteAccount(account, idx)">
                        <i class="fa fas fa-fw fa-trash-alt"></i>
                    </a>
                </li>
            </ul>
            <div class="setting-group-footer">
                <a @click="loadAccounts">Refresh accounts</a>
                <a v-if="activeAccount.owner" @click="showDialogNewAccount">Add new account</a>
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
		customDialog,
	},
	data() {
		return {
			loading: false,
			accounts: [],
		};
	},
	methods: {
		saveSetting() {
			let options = {
				ShowId: this.appOptions.ShowId,
				ListMode: this.appOptions.ListMode,
				HideThumbnail: this.appOptions.HideThumbnail,
				HideExcerpt: this.appOptions.HideExcerpt,
				NightMode: this.appOptions.NightMode,
			};

			if (this.activeAccount.owner) {
				options = {
					...options,
					KeepMetadata: this.appOptions.KeepMetadata,
					UseArchive: this.appOptions.UseArchive,
					CreateEbook: this.appOptions.CreateEbook,
					MakePublic: this.appOptions.MakePublic,
				};
			}

			this.$emit("setting-changed", options);
			//request
			fetch(new URL("/api/v1/auth/account", document.baseURI), {
				method: "PATCH",
				body: JSON.stringify({
					config: this.appOptions,
				}),
				headers: {
					"Content-Type": "application/json",
					Authorization: "Bearer " + localStorage.getItem("shiori-token"),
				},
			})
				.then((response) => {
					if (!response.ok) throw response;
					return response.json();
				})
				.then((responseData) => {
					const responseString = JSON.stringify(responseData.message);
					localStorage.setItem("shiori-account", responseString);
				})
				.catch((err) => {
					this.getErrorMessage(err).then((msg) => {
						this.showErrorDialog(msg);
					});
				});
		},
		loadAccounts() {
			if (this.loading) return;

			this.loading = true;
			fetch(new URL("api/accounts", document.baseURI), {
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
					this.loading = false;
					this.accounts = json;
				})
				.catch((err) => {
					this.loading = false;
					this.getErrorMessage(err).then((msg) => {
						this.showErrorDialog(msg);
					});
				});
		},
		showDialogNewAccount() {
			this.showDialog({
				title: "New Account",
				content: "Input new account's data :",
				fields: [
					{
						name: "username",
						label: "Username",
						value: "",
					},
					{
						name: "password",
						label: "Password",
						type: "password",
						value: "",
					},
					{
						name: "repeat",
						label: "Repeat password",
						type: "password",
						value: "",
					},
					{
						name: "visitor",
						label: "This account is for visitor",
						type: "check",
						value: false,
					},
				],
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
					};

					this.dialog.loading = true;
					fetch(new URL("api/accounts", document.baseURI), {
						method: "post",
						body: JSON.stringify(request),
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
							this.dialog.loading = false;
							this.dialog.visible = false;

							this.accounts.push({
								username: data.username,
								owner: !data.visitor,
							});
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
		showDialogChangePassword(account) {
			this.showDialog({
				title: "Change Password",
				content: "Input new password :",
				fields: [
					{
						name: "oldPassword",
						label: "Old password",
						type: "password",
						value: "",
					},
					{
						name: "password",
						label: "New password",
						type: "password",
						value: "",
					},
					{
						name: "repeat",
						label: "Repeat password",
						type: "password",
						value: "",
					},
				],
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
					};

					this.dialog.loading = true;
					fetch(new URL("api/accounts", document.baseURI), {
						method: "put",
						body: JSON.stringify(request),
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
							this.dialog.loading = false;
							this.dialog.visible = false;
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
							Authorization: "Bearer " + localStorage.getItem("shiori-token"),
						},
					})
						.then((response) => {
							if (!response.ok) throw response;
							return response;
						})
						.then(() => {
							this.dialog.loading = false;
							this.dialog.visible = false;
							this.accounts.splice(idx, 1);
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
	},
	mounted() {
		this.loadAccounts();
	},
};
