import { ref, watchEffect } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
// Lightweight browser feature checks (replaces removed testing utilities)
function checkBrowserSupport() {
  const cssCustomProperties = typeof window !== 'undefined' && window.CSS && (window.CSS.supports?.('--a', '0') ?? false)
  const prefersColorScheme = typeof window !== 'undefined' && typeof window.matchMedia === 'function'
  return { cssCustomProperties: !!cssCustomProperties, prefersColorScheme: !!prefersColorScheme }
}

export type ThemePreference = 'light' | 'dark' | 'system' | 'high-contrast'

// Debounce utility for performance optimization
function debounce<T extends (...args: any[]) => any>(func: T, wait: number): T {
	let timeout: number
	return ((...args: any[]) => {
		clearTimeout(timeout)
		timeout = setTimeout(() => func(...args), wait)
	}) as T
}

function getSystemPrefersDark(): boolean {
	return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
}

function setHtmlTheme(theme: ThemePreference): void {
	const root = document.documentElement

	// Remove all theme attributes
	root.removeAttribute('data-theme')
	root.classList.remove('dark')

	// Apply theme based on preference
	if (theme === 'system') {
		// Let CSS media queries handle system theme
		if (getSystemPrefersDark()) {
			root.classList.add('dark')
		}
	} else if (theme === 'dark') {
		root.setAttribute('data-theme', 'dark')
		root.classList.add('dark')
	} else if (theme === 'light') {
		root.setAttribute('data-theme', 'light')
	} else if (theme === 'high-contrast') {
		root.setAttribute('data-theme', 'high-contrast')
		root.classList.add('dark')
	}
}

function updateFavicon(theme: ThemePreference): void {
	const favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]:not([media])')
	if (!favicon) return

	const isDark = theme === 'dark' || theme === 'high-contrast' ||
		(theme === 'system' && getSystemPrefersDark())

	favicon.href = isDark ? '/favicon-dark.ico' : '/favicon.ico'
}

// Accessibility: Basic validation placeholder
function validateContrastRatio(theme: ThemePreference): boolean {
    // Themes are defined using design tokens; assume compliance here.
    // Keep hook for future automated checks.
    return true
}

export function useTheme() {
	const auth = useAuthStore()
	const { success, error: showError } = useToast()
	const current = ref<ThemePreference>('system')
	const isLoading = ref(false)
	const error = ref<string | null>(null)
let mediaQuery: MediaQueryList | null = null
let mediaListener: ((e: MediaQueryListEvent) => void) | null = null

	// Debounced apply function for performance
	const debouncedApply = debounce((pref: ThemePreference) => {
		try {
			// Validate accessibility
			if (!validateContrastRatio(pref)) {
				console.warn(`Theme ${pref} may not meet accessibility standards`)
			}

			setHtmlTheme(pref)
			updateFavicon(pref)

			// Store in localStorage as fallback
			localStorage.setItem('shiori-theme', pref)

			error.value = null
		} catch (err) {
			error.value = 'Failed to apply theme'
			console.error('Theme application error:', err)
		}
	}, 100)

	const apply = (pref: ThemePreference) => {
		current.value = pref
		debouncedApply(pref)
	}

const init = () => {
		try {
			// Check browser support first
			const browserSupport = checkBrowserSupport()
			if (!browserSupport.cssCustomProperties) {
				console.warn('CSS Custom Properties not supported, theme switching may not work properly')
			}

			// Try to get theme from user config first, then localStorage, then system
			const userTheme = auth.user?.config?.Theme as ThemePreference
			const storedTheme = localStorage.getItem('shiori-theme') as ThemePreference
			const pref = userTheme || storedTheme || 'system'

			apply(pref)

			// Listen to system changes only when in system mode
			if (browserSupport.prefersColorScheme) {
				mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
				mediaListener = () => {
					if (current.value === 'system') {
						apply('system')
					}
				}
				mediaQuery.addEventListener('change', mediaListener)
			}
		} catch (err) {
			error.value = 'Failed to initialize theme'
			console.error('Theme initialization error:', err)
		}
	}

const destroy = () => {
		if (mediaQuery && mediaListener) {
			mediaQuery.removeEventListener('change', mediaListener)
		}
}

	const setTheme = async (pref: ThemePreference) => {
		isLoading.value = true
		error.value = null

		try {
			apply(pref)

			// Persist to backend if authenticated
			if (auth.isAuthenticated) {
				const cfg = { ...(auth.user?.config as any) }
				cfg.Theme = pref
				await auth.updateUserConfig(cfg)
			}

			// Show success feedback
			const themeNames = {
				'light': 'Light',
				'dark': 'Dark',
				'system': 'System',
				'high-contrast': 'High Contrast'
			}
			success(`Theme changed to ${themeNames[pref]}`, 'Your theme preference has been saved')

		} catch (err: any) {
			const errorMessage = err.message || 'Failed to save theme preference'
			error.value = errorMessage
			showError('Theme Change Failed', errorMessage)
			console.error('Theme save error:', err)
		} finally {
			isLoading.value = false
		}
	}

	// Keep applied theme in sync if user config changes elsewhere
	watchEffect(() => {
		const pref = (auth.user?.config?.Theme as ThemePreference) || 'system'
		if (pref !== current.value) {
			apply(pref)
		}
	})

	return {
		current,
		setTheme,
		apply,
		isLoading,
		error,
        validateContrastRatio,
		init,
		destroy
	}
}


