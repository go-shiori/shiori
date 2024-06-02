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
					NightMode: false,
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
		showDialog(cfg) {
			const base = this.defaultDialog();
			base.visible = true;
			for (const key in cfg) {
				if (Object.prototype.hasOwnProperty.call(cfg, key)) {
					base[key] = cfg[key];
				}
			}
			this.dialog = base;
		},
		async getErrorMessage(err) {
			switch (err.constructor) {
				case Error:
					return err.message;
				case Response: {
					const text = await err.text();
					return `${text} (${err.status})`;
				}
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
		showErrorDialog(msg) {
			const sessionError = this.isSessionError(msg),
				dialogContent = sessionError
					? "Session has expired, please login again."
					: msg;

			this.showDialog({
				visible: true,
				title: "Error",
				content: dialogContent,
				mainText: "OK",
				mainClick: () => {
					this.dialog.visible = false;
					if (sessionError) {
						const loginUrl = new Url("login", document.baseURI);
						loginUrl.query.dst = window.location.href;

						document.cookie = `session-id=; Path=${
							new URL(document.baseURI).pathname
						}; Expires=Thu, 01 Jan 1970 00:00:00 GMT;`;
						location.href = loginUrl.toString();
					}
				},
			});
		},
	},
};
