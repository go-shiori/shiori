import { apiRequest } from "../utils/api.js";

export default {
	props: {
		activeAccount: {
			type: Object,
			default() {
				return {
					id: 0,
					username: "",
					owner: false,
				};
			},
		},
		appOptions: {
			type: Object,
			default() {
				return {
					ShowId: false,
					ListMode: false,
					HideThumbnail: false,
					HideExcerpt: false,
					KeepMetadata: false,
					UseArchive: false,
					CreateEbook: false,
					MakePublic: false,
				};
			},
		},
	},
	data() {
		return {
			dialog: {},
		};
	},
	methods: {
		defaultDialog() {
			return {
				visible: false,
				loading: false,
				title: "",
				content: "",
				fields: [],
				showLabel: false,
				mainText: "Yes",
				secondText: "",
				mainClick: () => {
					this.dialog.visible = false;
				},
				secondClick: () => {
					this.dialog.visible = false;
				},
				escPressed: () => {
					if (!this.loading) this.dialog.visible = false;
				},
			};
		},
		showDialog(opt) {
			this.dialog = {
				visible: true,
				...opt,
			};
		},
		async getErrorMessage(err) {
			switch (err.constructor) {
				case Error:
					return err.message;
				case Response:
					var text = await err.text();

					// Handle new error messages
					if (text[0] == "{") {
						var json = JSON.parse(text);
						return json.error;
					}

					return `${text} (${err.status})`;
				default:
					return err;
			}
		},
		isSessionError(err) {
			switch (
				err
					.toString()
					.replace(/\(\d+\)/g, "")
					.trim()
					.toLowerCase()
			) {
				case "session is not exist":
				case "session has been expired":
					return true;
				default:
					return false;
			}
		},
		themeSwitch(theme) {
			switch (theme) {
				case "light":
					document.body.classList.remove("dark");
					document.body.classList.add("light");
					break;
				case "dark":
					document.body.classList.remove("light");
					document.body.classList.add("dark");
					break;
				case "follow":
					document.body.classList.remove("light", "dark");
					break;
				default:
					console.error("Invalid theme selected");
			}
		},
		showErrorDialog(msg) {
			this.showDialog({
				title: "Error",
				content: msg,
				mainText: "OK",
				mainClick: () => {
					this.dialog.visible = false;
				},
				escPressed: () => {
					this.dialog.visible = false;
				},
			});
		},
		async saveSetting(key, value) {
			try {
				await apiRequest(new URL("api/v1/settings", document.baseURI), {
					method: "PUT",
					body: JSON.stringify({ [key]: value }),
				});
			} catch (err) {
				this.showErrorDialog(err.message);
			}
		},
		async logout() {
			try {
				await apiRequest(new URL("api/v1/auth/logout", document.baseURI), {
					method: "POST",
				});

				// Clear session data
				document.cookie = `token=; Path=${
					new URL(document.baseURI).pathname
				}; Expires=Thu, 01 Jan 1970 00:00:00 GMT;`;

				localStorage.removeItem("shiori-account");
				localStorage.removeItem("shiori-token");

				// Reload page
				window.location.reload();
			} catch (err) {
				this.showErrorDialog(err.message);
			}
		},
	},
};
