<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTagsStore } from '@/stores/tags'
import { useToast } from '@/composables/useToast'
import { XIcon, PlusIcon } from '@/components/icons'

// Debounce utility function
const debounce = (func: Function, wait: number) => {
    let timeout: number
    return function executedFunction(...args: any[]) {
        const later = () => {
            clearTimeout(timeout)
            func(...args)
        }
        clearTimeout(timeout)
        timeout = setTimeout(later, wait) as any
    }
}

interface Tag {
    id: number
    name: string
}

// Type guard to check if a tag has valid data
const isValidTag = (tag: any): tag is Tag => {
    return tag && typeof tag.id === 'number' && typeof tag.name === 'string'
}

interface Props {
    modelValue: number[]
    disabled?: boolean
    placeholder?: string
}

interface Emits {
    (e: 'update:modelValue', value: number[]): void
}

const props = withDefaults(defineProps<Props>(), {
    disabled: false,
    placeholder: 'Add tags...'
})

const emit = defineEmits<Emits>()

const tagsStore = useTagsStore()
const { t } = useI18n()
const { success, error } = useToast()

// State
const inputValue = ref('')
const isOpen = ref(false)
const selectedTagIds = ref<number[]>([...props.modelValue])
const inputRef = ref<HTMLInputElement>()
const selectedIndex = ref(-1) // -1 means no selection, 0+ means index in suggestions
const searchResults = ref<Tag[]>([])
const isSearching = ref(false)

// Computed
const availableTags = computed(() => {
    if (!inputValue.value.trim()) return []

    const searchTerm = inputValue.value.toLowerCase()
    return searchResults.value.filter(tag =>
        isValidTag(tag) &&
        tag.name.toLowerCase().includes(searchTerm) &&
        !selectedTagIds.value.includes(tag.id)
    )
})

const selectedTags = computed(() => {
    return tagsStore.tags.filter(tag =>
        isValidTag(tag) && selectedTagIds.value.includes(tag.id)
    )
})

const showCreateOption = computed(() => {
    if (!inputValue.value.trim()) return false

    const searchTerm = inputValue.value.toLowerCase()
    const exactMatch = searchResults.value.some(tag =>
        isValidTag(tag) && tag.name.toLowerCase() === searchTerm
    )

    return !exactMatch && !selectedTagIds.value.some(id => {
        const tag = searchResults.value.find(t => isValidTag(t) && t.id === id)
        return tag && tag.name && tag.name.toLowerCase() === searchTerm
    })
})

// All suggestions including create option
const allSuggestions = computed(() => {
    const suggestions: Array<{ type: 'tag', tag: Tag } | { type: 'create', name: string }> = []

    // Add available tags
    availableTags.value.forEach(tag => {
        if (isValidTag(tag)) {
            suggestions.push({ type: 'tag', tag })
        }
    })

    // Add create option if needed
    if (showCreateOption.value) {
        suggestions.push({ type: 'create', name: inputValue.value.trim() })
    }

    return suggestions
})

// Current selection
const currentSelection = computed(() => {
    if (selectedIndex.value >= 0 && selectedIndex.value < allSuggestions.value.length) {
        return allSuggestions.value[selectedIndex.value]
    }
    return null
})

// Methods
const addTag = async (tag: Tag) => {
    if (selectedTagIds.value.includes(tag.id)) return

    selectedTagIds.value.push(tag.id)
    emit('update:modelValue', [...selectedTagIds.value])

    inputValue.value = ''
    isOpen.value = false
    selectedIndex.value = -1
    await nextTick()
    inputRef.value?.focus()
}

const createAndAddTag = async (tagName: string) => {
    try {
        const newTag = await tagsStore.createTag(tagName)
        if (isValidTag(newTag)) {
            // Add the new tag to search results so it's available for future searches
            searchResults.value.push(newTag)
            await addTag(newTag)

            // Show success toast
            success(
                t('tags.toast.created_success'),
                t('tags.toast.created_success_message')
            )
        }
    } catch (err) {
        console.error('Failed to create tag:', err)

        // Show error toast
        error(
            t('tags.toast.created_error'),
            t('tags.toast.created_error_message')
        )
    }
}

const removeTag = (tagId: number) => {
    selectedTagIds.value = selectedTagIds.value.filter(id => id !== tagId)
    emit('update:modelValue', [...selectedTagIds.value])
}

const handleInput = () => {
    isOpen.value = inputValue.value.trim().length > 0
    selectedIndex.value = -1 // Reset selection when typing
    debouncedSearch(inputValue.value) // Trigger debounced search
}

// Debounced search function
const debouncedSearch = debounce(async (searchTerm: string) => {
    if (!searchTerm.trim()) {
        // If no search term, load all tags
        try {
            const allTags = await tagsStore.fetchTags({ limit: 1000 })
            searchResults.value = allTags.filter(isValidTag)
        } catch (err) {
            console.warn('Failed to load all tags:', err)
            searchResults.value = []
        }
        return
    }

    isSearching.value = true
    try {
        const results = await tagsStore.searchTags(searchTerm)
        searchResults.value = results.filter(isValidTag)
    } catch (err) {
        console.warn('Failed to search tags:', err)
        searchResults.value = []
    } finally {
        isSearching.value = false
    }
}, 300) // 300ms delay

const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key === 'Backspace' && !inputValue.value && selectedTagIds.value.length > 0) {
        // Remove last selected tag
        selectedTagIds.value.pop()
        emit('update:modelValue', [...selectedTagIds.value])
    } else if (event.key === 'ArrowDown') {
        event.preventDefault()
        if (!isOpen.value && inputValue.value.trim()) {
            isOpen.value = true
        }
        if (allSuggestions.value.length > 0) {
            selectedIndex.value = Math.min(selectedIndex.value + 1, allSuggestions.value.length - 1)
        }
    } else if (event.key === 'ArrowUp') {
        event.preventDefault()
        if (allSuggestions.value.length > 0) {
            selectedIndex.value = Math.max(selectedIndex.value - 1, -1)
        }
    } else if (event.key === 'Enter') {
        event.preventDefault()
        if (currentSelection.value) {
            if (currentSelection.value.type === 'tag') {
                await addTag(currentSelection.value.tag)
            } else if (currentSelection.value.type === 'create') {
                await createAndAddTag(currentSelection.value.name)
            }
        } else if (showCreateOption.value && inputValue.value.trim()) {
            await createAndAddTag(inputValue.value.trim())
        } else if (availableTags.value.length === 1) {
            const tag = availableTags.value[0]
            if (isValidTag(tag)) {
                await addTag(tag)
            }
        }
    } else if (event.key === 'Escape') {
        isOpen.value = false
        selectedIndex.value = -1
        inputRef.value?.blur()
    }
}

const handleFocus = () => {
    if (inputValue.value.trim()) {
        isOpen.value = true
    }
}

const handleBlur = () => {
    // Delay closing to allow for clicks on suggestions
    setTimeout(() => {
        isOpen.value = false
    }, 200)
}

// Watch for external changes
watch(() => props.modelValue, (newValue) => {
    selectedTagIds.value = [...newValue]
}, { deep: true })

// Load tags on mount
onMounted(async () => {
    try {
        // Load initial tags for better UX
        const allTags = await tagsStore.fetchTags({ limit: 1000 })
        searchResults.value = allTags.filter(isValidTag)
    } catch (err) {
        console.warn('Failed to load initial tags:', err)
    }
})
</script>

<template>
    <div class="relative">
        <!-- Selected Tags Display -->
        <div
            class="min-h-[42px] p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 flex flex-wrap gap-2 items-center">
            <!-- Selected Tag Chips -->
            <div v-for="tag in selectedTags" :key="tag.id"
                class="inline-flex items-center px-2 py-1 rounded-md text-sm bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200">
                <span>{{ tag.name }}</span>
                <button v-if="!disabled" @click="removeTag(tag.id!)"
                    class="ml-1 hover:text-red-600 dark:hover:text-red-400" type="button">
                    <XIcon class="h-3 w-3" />
                </button>
            </div>

            <!-- Input Field -->
            <input ref="inputRef" v-model="inputValue" :placeholder="selectedTags.length === 0 ? placeholder : ''"
                :disabled="disabled"
                class="flex-1 min-w-[120px] bg-transparent border-none outline-none text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                @input="handleInput" @keydown="handleKeydown" @focus="handleFocus" @blur="handleBlur" />
        </div>

        <!-- Dropdown Suggestions -->
        <div v-if="isOpen && (allSuggestions.length > 0 || isSearching)"
            class="absolute z-50 w-full mt-1 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg max-h-60 overflow-auto">

            <!-- Loading indicator -->
            <div v-if="isSearching" class="w-full px-3 py-2 text-center text-gray-500 dark:text-gray-400">
                <div class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-red-500 mr-2"></div>
                {{ t('tags.searching') }}
            </div>

            <!-- All Suggestions -->
            <button v-for="(suggestion, index) in allSuggestions"
                :key="suggestion.type === 'tag' ? suggestion.tag.id : `create-${suggestion.name}`"
                @click="suggestion.type === 'tag' ? addTag(suggestion.tag) : createAndAddTag(suggestion.name)" :class="[
                    'w-full px-3 py-2 text-left focus:outline-none',
                    index === selectedIndex
                        ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200'
                        : 'text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700',
                    suggestion.type === 'create' && index > 0 ? 'border-t border-gray-200 dark:border-gray-600' : ''
                ]" type="button">
                <span v-if="suggestion.type === 'tag'">{{ suggestion.tag.name }}</span>
                <span v-else class="text-red-600 dark:text-red-400 font-medium flex items-center space-x-2">
                    <PlusIcon size="14" />
                    <span>Create tag "{{ suggestion.name }}"</span>
                </span>
            </button>
        </div>
    </div>
</template>
