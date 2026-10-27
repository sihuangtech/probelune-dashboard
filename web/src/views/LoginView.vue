<script setup lang="ts">
import { reactive, ref } from 'vue'
import { api, user } from '../lib/api'
import type { User } from '../lib/models'
import { t } from '../lib/i18n'
const props = defineProps<{ initialized: boolean }>()
const emit = defineEmits<{ signedIn: [] }>()
const form = reactive({ username: '', password: '', setupToken: '' })
const busy = ref(false),
  error = ref('')
async function submit() {
  busy.value = true
  error.value = ''
  try {
    const data = await api<{ user: User }>(
      props.initialized ? '/auth/login' : '/auth/setup',
      'POST',
      { username: form.username, password: form.password },
      { 'X-Setup-Token': form.setupToken },
    )
    user.value = data.user
    emit('signedIn')
  } catch (e) {
    error.value = String(e)
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <div class="login-layout">
    <section class="login-story">
      <div class="brand"><span class="brand-orbit">◒</span>Probelune</div>
      <span class="eyebrow">DISTRIBUTED OBSERVABILITY</span>
      <h1>{{ t('tagline') }}</h1>
      <div class="orbital-map">
        <span class="orbit o1" /><span class="orbit o2" /><span class="orbit o3" /><b>巡汐</b>
      </div>
      <p>{{ t('nodeHint') }}</p>
    </section>
    <section class="login-form">
      <p class="eyebrow">YOUR PRIVATE CONTROL ROOM</p>
      <h2>{{ t('connect') }}</h2>
      <p class="muted">{{ t(initialized ? 'login' : 'secureHint') }}</p>
      <el-form label-position="top" @submit.prevent="submit"
        ><el-form-item :label="t('username')"
          ><el-input v-model="form.username" autocomplete="username" /></el-form-item
        ><el-form-item :label="t('password')"
          ><el-input
            v-model="form.password"
            type="password"
            show-password
            :autocomplete="initialized ? 'current-password' : 'new-password'" /></el-form-item
        ><el-form-item v-if="!initialized" :label="t('setupToken')"
          ><el-input v-model="form.setupToken" type="password" show-password /></el-form-item
        ><el-alert v-if="error" :title="error" type="error" :closable="false" /><el-button
          class="full-width"
          type="primary"
          native-type="submit"
          :loading="busy"
          >{{ t(initialized ? 'login' : 'setup') }}</el-button
        ></el-form
      >
    </section>
  </div>
</template>
