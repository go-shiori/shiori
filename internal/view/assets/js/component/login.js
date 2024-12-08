export default {
    data() {
        return {
            error: "",
            loading: false,
            username: "",
            password: "",
            remember: false,
        }
    },
    methods: {
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
        parseJWT(token) {
            try {
                return JSON.parse(atob(token.split('.')[1]));
            } catch (e) {
                return null;
            }
        },
        login() {
            // needed to work around autofill issue
            // https://github.com/facebook/react/issues/1159#issuecomment-506584346
            this.username = document.querySelector('#username').value;
            this.password = document.querySelector('#password').value;
            
            // Validate input
            if (this.username === "") {
                this.error = "Username must not empty";
                return;
            }

            // Remove old cookie
            document.cookie = `session-id=; Path=${new URL(document.baseURI).pathname}; Expires=Thu, 01 Jan 1970 00:00:00 GMT;`;

            // Send request
            this.loading = true;

            fetch(new URL("api/v1/auth/login", document.baseURI), {
                method: "post",
                body: JSON.stringify({
                    username: this.username,
                    password: this.password,
                    remember_me: this.remember == 1 ? true : false,
                }),
                headers: { "Content-Type": "application/json" },
            }).then(response => {
                if (!response.ok) throw response;
                return response.json();
            }).then(json => {
                // Save session id
                document.cookie = `session-id=${json.message.session}; Path=${new URL(document.baseURI).pathname}; Expires=${json.message.expires}`;
                document.cookie = `token=${json.message.token}; Path=${new URL(document.baseURI).pathname}; Expires=${json.message.expires}`;

                // Save account data
                localStorage.setItem("shiori-token", json.message.token);
                localStorage.setItem("shiori-account", JSON.stringify(this.parseJWT(json.message.token).account));

                // Go to destination page
                var currentUrl = new Url(),
                    dstUrl = currentUrl.query.dst,
                    dstPage = currentUrl.hash || "home";

                if (dstPage !== "home" && dstPage !== "setting") {
                    dstPage = "";
                }

                // TODO: Improve this redirect logic
                var newUrl = new Url(dstUrl || document.baseURI);
                newUrl.hash = dstPage;

                location.href = newUrl;
            }).catch(err => {
                this.loading = false;
                this.getErrorMessage(err).then(msg => {
                    this.error = msg;
                })
            });
        }
    },
    mounted() {
        // Load setting
        localStorage.removeItem("shiori-account");
        localStorage.removeItem("shiori-token");

        // <input autofocus> wasn't working all the time, so I'm putting this here as a fallback
        document.querySelector('#username').focus()
    }
}
