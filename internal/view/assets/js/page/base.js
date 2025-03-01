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
		showDialog(cfg) {
			var base = this.defaultDialog();
			base.visible = true;
			if (cfg.loading) base.loading = cfg.loading;
			if (cfg.title) base.title = cfg.title;
			if (cfg.content) base.content = cfg.content;
			if (cfg.fields) base.fields = cfg.fields;
			if (cfg.showLabel) base.showLabel = cfg.showLabel;
			if (cfg.mainText) base.mainText = cfg.mainText;
			if (cfg.secondText) base.secondText = cfg.secondText;
			if (cfg.mainClick) base.mainClick = cfg.mainClick;
			if (cfg.secondClick) base.secondClick = cfg.secondClick;
			if (cfg.escPressed) base.escPressed = cfg.escPressed;
			this.dialog = base;
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
						return json.message;
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
			var sessionError = this.isSessionError(msg),
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
						var loginUrl = new Url("login", document.baseURI);
						loginUrl.query.dst = window.location.href;

						document.cookie = `token=; Path=${
							new URL(document.baseURI).pathname
						}; Expires=Thu, 01 Jan 1970 00:00:00 GMT;`;
						location.href = loginUrl.toString();
					}
				},
			});
		},
	},
};
