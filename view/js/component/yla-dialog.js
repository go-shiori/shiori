var YlaDialog = function () {
    // Private variable
    var _template = `
        <div v-if="visible" class="yla-dialog__overlay">
            <div class="yla-dialog">
                <div class="yla-dialog__header">
                    <p>{{title}}</p>
                </div>
                <div class="yla-dialog__body">
                    <slot>
                        <p class="yla-dialog__content">{{content}}</p>
                        <template v-for="(field,index) in formFields">
                            <p v-if="showLabel">{{field.label}} :</p>
                            <input :style="{gridColumnEnd: showLabel ? null : 'span 2'}" 
                                   :type="fieldType(field)" 
                                   :placeholder="field.label" 
                                   :tabindex="index+1"
                                   ref="input"
                                   v-model="field.value" 
                                   @focus="$event.target.select()"
                                   @keyup="handleInput(index)"
                                   @keyup.enter="handleInputEnter(index)">
                            <span ref="suggestion" v-if="field.suggestion" class="suggestion">{{field.suggestion}}</span>
                        </template>
                    </slot>
                </div>
                <div class="yla-dialog__footer">
                    <i v-if="loading" class="fas fa-fw fa-spinner fa-spin"></i>
                    <slot v-else name="custom-footer">
                        <a v-if="secondText" 
                           :tabindex="btnTabIndex+1" 
                           @click="handleSecondClick" 
                           @keyup.enter="handleSecondClick" 
                           class="yla-dialog__button">{{secondText}}
                        </a>
                        <a :tabindex="btnTabIndex" 
                           ref="mainButton"
                           @click="handleMainClick" 
                           @keyup.enter="handleMainClick" 
                           class="yla-dialog__button main">{{mainText}}
                        </a>
                    </slot>
                </div>
            </div>
        </div>`;

    return {
        template: _template,
        props: {
            visible: Boolean,
            loading: Boolean,
            title: {
                type: String,
                default: ''
            },
            content: {
                type: String,
                default: ''
            },
            fields: {
                type: Array,
                default () {
                    return []
                }
            },
            showLabel: {
                type: Boolean,
                default: false
            },
            mainText: {
                type: String,
                default: 'OK'
            },
            secondText: String,
            mainClick: {
                type: Function,
                default () {}
            },
            secondClick: {
                type: Function,
                default () {}
            }
        },
        data() {
            return {
                formFields: []
            };
        },
        computed: {
            btnTabIndex() {
                return this.fields.length + 1;
            }
        },
        watch: {
            fields: {
                immediate: true,
                handler() {
                    this.formFields = this.fields.map(field => {
                        if (typeof field === 'string') return {
                            name: field,
                            label: field,
                            value: '',
                            type: 'text',
                            dictionary: [],
                            suggestion: undefined
                        }

                        if (typeof field === 'object') return {
                            name: field.name || '',
                            label: field.label || '',
                            value: field.value || '',
                            type: field.type || 'text',
                            dictionary: field.dictionary instanceof Array ? field.dictionary : [],
                            suggestion: undefined
                        }
                    });
                }
            },
            'fields.length' () {
                this.focus();
            },
            visible() {
                this.focus();
            }
        },
        methods: {
            fieldType(f) {
                var type = f.type || 'text';
                if (type !== 'text' && type !== 'password') return 'text';
                else return type;
            },
            handleMainClick() {
                var data = {};
                this.formFields.forEach(field => {
                    var value = field.value;
                    if (field.type === 'number') value = parseInt(value, 10) || 0;
                    else if (field.type === 'float') value = parseFloat(value) || 0.0;
                    data[field.name] = value;
                })
                this.mainClick(data);
            },
            handleSecondClick() {
                this.secondClick();
            },
            handleInput(index) {
                // Create initial variable
                var field = this.formFields[index],
                    dictionary = field.dictionary;

                // Make sure dictionary is not empty
                if (dictionary.length === 0) return;

                // Fetch suggestion from dictionary
                var words = field.value.split(' '),
                    lastWord = words[words.length - 1].toLowerCase(),
                    suggestion;

                if (lastWord !== '') {
                    suggestion = dictionary.find(word => {
                        return word.toLowerCase().startsWith(lastWord)
                    });
                }

                this.formFields[index].suggestion = suggestion;

                // Make sure suggestion exist
                if (suggestion == null) return;

                // Display suggestion
                this.$nextTick(() => {
                    var input = this.$refs.input[index],
                        span = this.$refs.suggestion[index],
                        inputRect = input.getBoundingClientRect();

                    span.style.top = (inputRect.bottom - 1) + 'px';
                    span.style.left = inputRect.left + 'px';
                });
            },
            handleInputEnter(index) {
                var suggestion = this.formFields[index].suggestion;

                if (suggestion == null) {
                    this.handleMainClick();
                    return;
                }

                var words = this.formFields[index].value.split(' ');
                words.pop();
                words.push(suggestion);

                this.formFields[index].value = words.join(' ') + ' ';
                this.formFields[index].suggestion = undefined;
            },
            focus() {
                this.$nextTick(() => {
                    if (!this.visible) return;

                    var fields = this.$refs.input,
                        otherInput = this.$el.querySelectorAll('input'),
                        button = this.$refs.mainButton;

                    if (fields && fields.length > 0) {
                        this.$refs.input[0].focus();
                        this.$refs.input[0].select();
                    } else if (otherInput && otherInput.length > 0) {
                        otherInput[0].focus();
                        otherInput[0].select();
                    } else if (button) {
                        button.focus();
                    }
                });
            }
        }
    };
};