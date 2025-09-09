import { apiRequest } from "../utils/api.js";

const template = `
<div id="login-scene">
    <p class="error-message" v-if="error !== ''">{{error}}</p>
    <div id="login-box">
        <form @submit.prevent="login">
            <div id="logo-area">
                <p id="logo">
                    <span>栞</span>shiori
                </p>
                <p id="tagline">simple bookmark manager</p>
            </div>
            <div id="input-area">
                <label for="username">Username: </label>
                <input id="username" type="text" name="username" placeholder="Username" tabindex="1" autofocus />
                <label for="password">Password: </label>
                <input id="password" type="password" name="password" placeholder="Password" tabindex="2"
                    @keyup.enter="login">
                <label class="checkbox-field"><input type="checkbox" name="remember" v-model="remember"
                        tabindex="3">Remember me</label>
            </div>
            <div id="button-area">
                <a v-if="loading">
                    <i class="fas fa-fw fa-spinner fa-spin"></i>
                </a>
                <a v-else class="button" tabindex="4" @click="login" @keyup.enter="login">Log In</a>
            </div>
        </form>
    </div>
</div>
`;

export default {
	name: "login-view",
	template,
	data() {
		return {
			error: "",
			loading: false,
			username: "",
			password: "",
			remember: false,
			destination: "/", // Default destination
		};
	},
	emits: ["login-success"],
	methods: {
		sanitizeDestination(dst) {
			try {
				// Remove any leading/trailing whitespace
				dst = dst.trim();

				// Decode the URL to handle any encoded characters
				dst = decodeURIComponent(dst);

				// Create a URL object to parse the destination
				const url = new URL(dst, window.location.origin);

				// Only allow paths from the same origin
				if (url.origin !== window.location.origin) {
					return "/";
				}

				// Only return the pathname and search params
				return url.pathname + url.search + url.hash;
			} catch (e) {
				// If any error occurs during parsing, return root
				return "/";
			}
		},

		parseJWT(token) {
			try {
				return JSON.parse(atob(token.split(".")[1]));
			} catch (e) {
				return null;
			}
		},

		async login() {
			// Get values directly from the form
			const usernameInput = document.querySelector("#username");
			const passwordInput = document.querySelector("#password");
			this.username = usernameInput ? usernameInput.value : this.username;
			this.password = passwordInput ? passwordInput.value : this.password;

			// Validate input
			if (this.username === "") {
				this.error = "Username must not empty";
				return;
			}

			// Remove old cookie
			document.cookie = `token=; Path=${
				new URL(document.baseURI).pathname
			}; Expires=Thu, 01 Jan 1970 00:00:00 GMT;`;

			// Send request
			this.loading = true;

			try {
				const json = await apiRequest(
					new URL("api/v1/auth/login", document.baseURI),
					{
						method: "post",
						body: JSON.stringify({
							username: this.username,
							password: this.password,
							remember_me: this.remember == 1 ? true : false,
						}),
					},
				);

				// Save session id
				document.cookie = `token=${json.token}; Path=${
					new URL(document.baseURI).pathname
				}; Expires=${new Date(json.expires * 1000).toUTCString()}`;

				// Save account data
				localStorage.setItem("shiori-token", json.token);
				localStorage.setItem(
					"shiori-account",
					JSON.stringify(this.parseJWT(json.token).account),
				);

				this.visible = false;
				this.$emit("login-success");

				// Redirect to sanitized destination
				if (this.destination !== "/") window.location.href = this.destination;
			} catch (err) {
				this.error = err.message;
			} finally {
				this.loading = false;
			}
		},

		async checkSession() {
			const token = localStorage.getItem("shiori-token");
			if (!token) return false;

			try {
				await apiRequest(new URL("api/v1/auth/me", document.baseURI));
				return true;
			} catch (err) {
				return false;
			}
		},
	},
	async mounted() {
		// Get and sanitize destination from URL parameters
		const urlParams = new URLSearchParams(window.location.search);
		const dst = urlParams.get("dst");
		this.destination = dst ? this.sanitizeDestination(dst) : "/";

		// Check if there's a valid session
		if (await this.checkSession()) {
			this.$emit("login-success");
			return;
		}

		// Clear session data if we reach here
		document.cookie = `token=; Path=${
			new URL(document.baseURI).pathname
		}; Expires=Thu, 01 Jan 1970 00:00:00 GMT;`;

		localStorage.removeItem("shiori-account");
		localStorage.removeItem("shiori-token");

		// Focus username input
		this.$nextTick(() => {
			const usernameInput = document.querySelector("#username");
			if (usernameInput) {
				usernameInput.focus();
			}
		});
	},
};
