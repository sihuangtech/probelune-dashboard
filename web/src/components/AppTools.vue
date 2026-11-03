<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faGithub } from '@fortawesome/free-brands-svg-icons'
import { t, languageNames, languagePreference } from '../lib/i18n'
import { theme } from '../lib/theme'
</script>

<template>
  <div class="app-tools">
    <label
      ><span>{{ t('language') }}</span
      ><select
        class="language-toggle"
        v-model="languagePreference"
        :aria-label="t('language')"
        :title="t('language')"
      >
        <option value="system">{{ t('systemLanguage') }}</option>
        <option v-for="(name, tag) in languageNames" :key="tag" :value="tag">{{ name }}</option>
      </select></label
    >
    <label
      ><span>{{ t('theme') }}</span
      ><select v-model="theme" class="theme-toggle" :aria-label="t('theme')" :title="t('theme')">
        <option v-for="mode in ['light', 'dark', 'system']" :key="mode" :value="mode">
          {{ t('theme' + mode[0]!.toUpperCase() + mode.slice(1)) }}
        </option>
      </select></label
    >
    <a
      class="github-link"
      href="https://github.com/sihuangtech/probelune-dashboard"
      target="_blank"
      rel="noopener noreferrer"
      :aria-label="t('sourceCode')"
      :title="t('sourceCode')"
    >
      <FontAwesomeIcon :icon="faGithub" />
    </a>
  </div>
</template>

<style scoped>
label {
  min-width: 0;
}
label span {
  display: block;
  font-size: 11px;
  color: var(--muted);
}
.github-link {
  align-self: end;
}
</style>
