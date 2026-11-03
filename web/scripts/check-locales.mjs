import { readFileSync, readdirSync } from 'node:fs'
import assert from 'node:assert/strict'
import { createI18n } from 'vue-i18n'
const directory = new URL('../src/lib/locales/', import.meta.url)
const messages = Object.fromEntries(readdirSync(directory).filter(name => name.endsWith('.json')).map(name => [name.slice(0,-5), JSON.parse(readFileSync(new URL(name,directory),'utf8'))]))
const expected = Object.keys(messages.en).sort()
for (const [locale, catalog] of Object.entries(messages)) {
  assert.deepEqual(Object.keys(catalog).sort(),expected,locale)
  const i18n = createI18n({legacy:false, locale, messages:{[locale]:catalog}})
  for(const key of expected) {
    assert.equal(typeof catalog[key],'string',locale + ' ' + key)
    assert.ok(catalog[key].trim(),locale + ' ' + key)
    assert.ok(i18n.global.t(key).trim(),locale + ' ' + key)
  }
  assert.equal(catalog.appName,locale.startsWith('zh')?'巡汐':'Probelune')
}
console.log('Validated eight complete locale catalogs and Vue I18n message rendering.')
