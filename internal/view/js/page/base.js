export default {
    props: {
        displayOptions: {
            type: Object,
            default () {
                return {
                    showId: false,
                    listMode: false,
                    nightMode: false,
                    useArchive: false,
                };
            }
        }
    },
    data() {
        return {
            dialog: {}
        }
    },
    methods: {
        defaultDialog() {
            return {
                visible: false,
                loading: false,
                title: '',
                content: '',
                fields: [],
                showLabel: false,
                mainText: 'Yes',
                secondText: '',
                mainClick: () => {
                    this.dialog.visible = false;
                },
                secondClick: () => {
                    this.dialog.visible = false;
                },
                escPressed: () => {
                    if (!this.loading) this.dialog.visible = false;
                }
            }
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
            this.dialog = base;
        },
        isSessionError(err) {
            switch (err.trim().toLowerCase()) {
                case "session is not exist":
                case "session has been expired":
                    return true
                default:
                    return false;
            }
        },
        showErrorDialog(msg, status) {
            var sessionError = this.isSessionError(msg),
                dialogContent = sessionError ? "Session has expired, please login again." : `${msg} (${status})`;

            this.showDialog({
                visible: true,
                title: 'Error',
                content: dialogContent,
                mainText: 'OK',
                mainClick: () => {
                    this.dialog.visible = false;
                    if (sessionError) {
                        document.cookie = "session-id=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;";
                        location.href = "/login";
                    }
                }
            });
        },
    }
}