export default {
    data() {
        return {
            dialog: {},
            displayOptions: {
                showId: false,
                listMode: true,
            }
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
        showErrorDialog(msg) {
            this.showDialog({
                visible: true,
                title: 'Error',
                content: msg,
                mainText: 'OK',
                mainClick: () => {
                    this.dialog.visible = false;
                }
            });
        }
    }
}