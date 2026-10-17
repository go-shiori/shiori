var template = `
<div id="page-setting">
    <h1 class="page-header">Settings</h1>
    <div class="setting-container">
        <details open class="setting-group" id="setting-display">
            <summary>Display</summary>
            <label>
                Theme &nbsp;
                <select v-model="appOptions.Theme" @change="saveSetting">
                <option value="follow">Follow system</option>
                <option value="light">Light theme</option>
                <option value="dark">Dark theme</option>
                </select>
            </label>
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
        </details>
        <details v-if="activeAccount.Owner" open class="setting-group" id="setting-bookmarks">
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
                Make bookmark publicly available by default
            </label>
        </details>
        <details v-if="activeAccount.Owner" open class="setting-group setting-accounts" id="setting-accounts">
            <summary>Accounts</summary>
            <ul class="accounts-list">
                <li v-if="accounts.length === 0">No accounts registered</li>
                <li v-for="(account, idx) in accounts" :shiori-username="account.Username">
                    <p>{{account.Username}}
                        <span v-if="account.Owner" class="account-level">(owner)</span>
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
                <a @click="loadAccounts" title="Refresh accounts">Refresh accounts</a>
                <a v-if="activeAccount.Owner" @click="showDialogNewAccount" title="Add new account">Add new account</a>
            </div>
        </details>
        <details v-if="!activeAccount.Owner" open class="setting-group setting-accounts" id="setting-my-account">
            <summary>My account</summary>
            <ul>
                <li v-for="(account, idx) in [this.activeAccount]" :shiori-username="account.Username">
                    <p>{{account.Username}}
                        <span v-if="account.Owner" class="account-level">(owner)</span>
                    </p>
                    <a title="Change password" @click="showDialogChangePassword(account)">
                        <i class="fa fas fa-fw fa-key"></i>
                    </a>
                </li>
            </ul>
            <div class="setting-group-footer">
                <a @click="showDialogChangePassword(this.activeAccount)" title="Change password">Change password</a>
            </div>
        </details>
		<details v-if="activeAccount.Owner" class="setting-group" id="setting-system-info">
			<summary>System info</summary>
			<ul>
				<li><b>Shiori version:</b> <span>{{system.Version?.Tag}}<span></li>
				<li><b>Database engine:</b> <span>{{system.Database}}</span></li>
				<li><b>Operating system:</b> <span>{{system.OS}}</span></li>
			</ul>
		</details>
    </div>
    <div class="loading-overlay" v-if="loading"><i class="fas fa-fw fa-spin fa-spinner"></i></div>
    <custom-dialog v-bind="dialog"/>
</div>`;

import customDialog from "../component/dialog.js";
import basePage from "./base.js";
import { apiRequest } from "../utils/api.js";

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
			system: {},
		};
	},
	methods: {
		saveSetting() {
			let options = {
				ShowId: this.appOptions.ShowId,
				ListMode: this.appOptions.ListMode,
				HideThumbnail: this.appOptions.HideThumbnail,
				HideExcerpt: this.appOptions.HideExcerpt,
				Theme: this.appOptions.Theme,
			};

			if (this.activeAccount.Owner) {
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
			fetch(new URL("api/v1/auth/account", document.baseURI), {
				method: "PATCH",
				body: JSON.stringify({
					Config: this.appOptions,
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
					const responseString = JSON.stringify(responseData);
					localStorage.setItem("shiori-account", responseString);
				})
				.catch((err) => {
					this.getErrorMessage(err).then((msg) => {
						this.showErrorDialog(msg);
					});
				});
		},
		async loadAccounts() {
			if (this.loading) return;

			this.loading = true;
			try {
				const json = await apiRequest(
					new URL("api/v1/accounts", document.baseURI),
				);
				this.loading = false;
				this.accounts = json;
			} catch (err) {
				this.loading = false;
				this.showErrorDialog(err.message);
			}
		},
		async loadSystemInfo() {
			if (this.system.Version !== undefined) return;

			try {
				const json = await apiRequest(
					new URL("api/v1/system/info", document.baseURI),
				);
				this.system = json;
			} catch (err) {
				this.showErrorDialog(err.message);
			}
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
						name: "repeat_password",
						label: "Repeat password",
						type: "password",
						value: "",
					},
					{
						name: "admin",
						label: "This account is an admin account",
						type: "check",
						value: false,
					},
				],
				mainText: "OK",
				secondText: "Cancel",
				mainClick: async (data) => {
					if (data.username === "") {
						return;
					}

					var request = {
						Username: data.username,
						Password: data.password,
						Owner: data.admin,
					};

					this.dialog.loading = true;
					try {
						const json = await apiRequest(
							new URL("api/v1/accounts", document.baseURI),
							{
								method: "post",
								body: JSON.stringify(request),
							},
						);

						this.dialog.loading = false;
						this.dialog.visible = false;

						this.accounts.push(json);
						this.accounts.sort((a, b) => {
							var nameA = a.Username.toLowerCase(),
								nameB = b.Username.toLowerCase();

							if (nameA < nameB) {
								return -1;
							}

							if (nameA > nameB) {
								return 1;
							}

							return 0;
						});
					} catch (err) {
						this.dialog.loading = false;
						this.showErrorDialog(err.message);
					}
				},
			});
		},
		showDialogChangePassword(account) {
			let fields = [
				{
					name: "new_password",
					label: "New password",
					type: "password",
					value: "",
				},
				{
					name: "repeat_password",
					label: "Repeat password",
					type: "password",
					value: "",
				},
			];

			const requiresOldPassword =
				!this.activeAccount.Owner || this.activeAccount.ID === account.ID;

			// Only owners can update user passwords without
			// providing the old password

			if (requiresOldPassword) {
				fields.unshift({
					name: "old_password",
					label: "The current password",
					type: "password",
					value: "",
				});
			}

			this.showDialog({
				title: "Change Password",
				content: "",
				fields: fields,
				mainText: "OK",
				secondText: "Cancel",
				mainClick: async (data) => {
					if (requiresOldPassword) {
						if (data.old_password === "") {
							this.showErrorDialog("You must provide the current password.");
							return;
						}
					}

					if (data.new_password === "") {
						this.showErrorDialog("New password must not empt.");
						return;
					}

					if (data.new_password !== data.repeat_password) {
						this.showErrorDialog("Password does not match.");
						return;
					}

					var request = {
						OldPassword: data.old_password,
						NewPassword: data.new_password,
					};

					// Determine which URL to use depending if the user is updating its own
					// account or another user's account.
					let url = `api/v1/accounts/${account.ID}`;
					if (this.activeAccount.ID === account.ID) {
						url = "api/v1/auth/account";
					}

					this.dialog.loading = true;
					try {
						await apiRequest(new URL(url, document.baseURI), {
							method: "PATCH",
							body: JSON.stringify(request),
						});

						this.showDialog({
							title: "Password Changed",
							content: "Password has been changed.",
							mainText: "OK",
							mainClick: () => {
								this.dialog.visible = false;
							},
						});
					} catch (err) {
						this.dialog.loading = false;
						this.showErrorDialog(err.message);
					}
				},
			});
		},
		showDialogDeleteAccount(account, idx) {
			this.showDialog({
				title: "Delete Account",
				content: `Delete account "${account.Username}" ?`,
				mainText: "Yes",
				secondText: "No",
				mainClick: async () => {
					this.dialog.loading = true;
					try {
						await apiRequest(`api/v1/accounts/${account.ID}`, {
							method: "DELETE",
						});

						this.dialog.loading = false;
						this.dialog.visible = false;
						this.accounts.splice(idx, 1);
					} catch (err) {
						this.dialog.loading = false;
						this.showErrorDialog(err.message);
					}
				},
			});
		},
	},
	mounted() {
		if (this.activeAccount.Owner) {
			this.loadAccounts();
			this.loadSystemInfo();
		}
	},
};
