var template = `
<div v-if="visible" class="custom-dialog-overlay" @keyup.esc="handleEscPressed">
	<div class="custom-dialog">
		<p class="custom-dialog-header">{{title}}</p>
		<div class="custom-dialog-body">
			<slot>
				<p class="custom-dialog-content">{{content}}</p>
				<template v-for="(field,index) in formFields">
					<label v-if="showLabel && field.type !== 'check'">{{field.label}} :</label>
					<textarea v-if="field.type === 'area'"
						:style="{gridColumnEnd: showLabel ? null : 'span 2'}" 
						:placeholder="field.label" 
						:tabindex="index+1"
						ref="input"
						v-model="field.value" 
						@focus="$event.target.select()"
						@keyup="handleInput(index)">
					</textarea>
					<label v-else-if="field.type === 'check'" class="checkbox-field">
						<input type="checkbox" 
							v-model="field.value" 
							:tabindex="index+1">{{field.label}}
					</label>
					<input v-else
						:style="{gridColumnEnd: showLabel ? null : 'span 2'}" 
						:type="fieldType(field)" 
						:placeholder="field.label" 
						:tabindex="index+1"
						ref="input"
						v-model="field.value" 
						@focus="$event.target.select()"
						@keyup="handleInput(index)"
						@keyup.enter="handleInputEnter(index)">
					<span :ref="'suggestion-'+index" 
						v-if="field.suggestion" 
						class="suggestion">{{field.suggestion}}</span>
				</template>
			</slot>
		</div>
		<div class="custom-dialog-footer">
			<i v-if="loading" class="fas fa-fw fa-spinner fa-spin"></i>
			<slot v-else name="custom-footer">
				<a v-if="secondText" 
					:tabindex="btnTabIndex+1" 
					@click="handleSecondClick" 
					@keyup.enter="handleSecondClick" 
					class="custom-dialog-button">{{secondText}}
				</a>
				<a :tabindex="btnTabIndex" 
					ref="mainButton"
					@click="handleMainClick" 
					@keyup.enter="handleMainClick" 
					class="custom-dialog-button main">{{mainText}}
				</a>
			</slot>
		</div>
	</div>
</div>`;

export default {
	template: template,
	props: {
		title: String,
		loading: Boolean,
		visible: Boolean,
		content: {
			type: String,
			default: ''
		},
		fields: {
			type: Array,
			default() {
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
			default() { this.visible = false; }
		},
		secondClick: {
			type: Function,
			default() { this.visible = false; }
		},
		escPressed: {
			type: Function,
			default() { this.visible = false; }
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
						separator: ' ',
						suggestion: undefined
					}

					if (typeof field === 'object') return {
						name: field.name || '',
						label: field.label || '',
						value: field.value || '',
						type: field.type || 'text',
						dictionary: field.dictionary instanceof Array ? field.dictionary : [],
						separator: field.separator || ' ',
						suggestion: undefined
					}
				});
			}
		},
		'fields.length'() {
			this.focus();
		},
		visible: {
			immediate: true,
			handler() { this.focus() }
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
				else if (field.type === 'check') value = Boolean(value);
				data[field.name] = value;
			})

			this.mainClick(data);
		},
		handleSecondClick() {
			this.secondClick();
		},
		handleEscPressed() {
			this.escPressed();
		},
		handleInput(index) {
			// Create initial variable
			var field = this.formFields[index],
				dictionary = field.dictionary;

			// Make sure dictionary is not empty
			if (dictionary.length === 0) return;

			// Fetch suggestion from dictionary
			var words = field.value.split(field.separator),
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
					span = this.$refs['suggestion-' + index][0],
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

			var separator = this.formFields[index].separator,
				words = this.formFields[index].value.split(separator);

			words.pop();
			words.push(suggestion);

			this.formFields[index].value = words.join(separator) + separator;
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
			})
		}
	}
}