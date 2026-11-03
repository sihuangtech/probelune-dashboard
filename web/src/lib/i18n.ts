import { ref, watch, watchEffect } from 'vue'
import { createI18n } from 'vue-i18n'
import backendErrors from './backendErrors.json'
import en from './locales/en.json'
import zhCN from './locales/zh-CN.json'
import zhTW from './locales/zh-TW.json'
import ja from './locales/ja.json'
import ko from './locales/ko.json'
import de from './locales/de.json'
import fr from './locales/fr.json'
import es from './locales/es.json'

export const languageNames: Record<string, string> = {
  en: 'English',
  'zh-CN': '简体中文',
  'zh-TW': '繁體中文',
  ja: '日本語',
  ko: '한국어',
  de: 'Deutsch',
  fr: 'Français',
  es: 'Español',
}
export function resolveLanguage(value: string) {
  const tag = value.toLowerCase()
  if (tag.startsWith('zh')) return /tw|hk|mo|hant/.test(tag) ? 'zh-TW' : 'zh-CN'
  const base = tag.split('-')[0]!
  return base in languageNames ? base : 'en'
}
export const i18n = createI18n({
  legacy: false,
  locale: resolveLanguage(
    localStorage.getItem('probelune-language') === 'system'
      ? navigator.language
      : localStorage.getItem('probelune-language') || navigator.language,
  ),
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en, 'zh-CN': zhCN, 'zh-TW': zhTW, ja, ko, de, fr, es },
})
export const language = i18n.global.locale
const saved = localStorage.getItem('probelune-language')
export const languagePreference = ref(saved === 'zh' ? 'zh-CN' : saved || 'system')
watch(
  languagePreference,
  (value) => {
    language.value = resolveLanguage(
      value === 'system' ? navigator.language : value,
    ) as typeof language.value
    localStorage.setItem('probelune-language', value)
  },
  { immediate: true },
)
export const t = (key: string) => i18n.global.t(key)
watchEffect(() => {
  document.documentElement.lang = language.value
  document.title = t('appName')
})

export function translateServerError(value: string) {
  const key = (backendErrors as Record<string, string>)[value]
  return key ? t(key) : value
}
