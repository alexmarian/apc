// src/utils/theme.ts
import { ref, computed } from 'vue'
// Don't import GlobalThemeOverrides directly from naive-ui
// Instead use a more generic type

// Create a symbol for the theme key for more reliable injection
export const ThemeKey = Symbol('theme')

// Theme options
export const themeOptions = ['light', 'dark', 'auto'] as const
export type ThemeOption = typeof themeOptions[number]

// Export the provider interface for TypeScript
export interface ThemeProvider {
  current: computed<any>
  overrides: any // Changed from GlobalThemeOverrides to any
  isDark: computed<boolean>
  switchTheme: (theme: ThemeOption) => void
  themeOptions: readonly string[]
}

// Create a default theme provider for fallback
export const defaultThemeProvider: ThemeProvider = {
  current: ref(null),
  overrides: {},
  isDark: ref(false),
  switchTheme: (theme: ThemeOption) => {
    console.error('Theme provider not found. Make sure ThemeProvider is a parent component.')
  },
  themeOptions
}
