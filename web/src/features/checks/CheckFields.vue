<script setup lang="ts">
import type { Check } from '../../lib/models'
import { t } from '../../lib/i18n'
import { useCheckCatalog } from './catalog'
const model = defineModel<Check>({ required: true })
defineProps<{ diagnostic?: boolean }>()
const { definitions, catalogError } = useCheckCatalog()
</script>
<template>
  <el-alert v-if="catalogError" :title="catalogError" type="error" :closable="false" />
  <div class="grid-two">
    <el-form-item :label="t('type')"
      ><el-select v-model="model.type"
        ><el-option
          v-for="kind in definitions.filter((item) => diagnostic || item.continuous)"
          :key="kind.type"
          :value="kind.type"
          :label="kind.type.toUpperCase()" /></el-select></el-form-item
    ><el-form-item :label="t('timeout')"
      ><el-input-number v-model="model.timeout" :min="1" :max="60"
    /></el-form-item>
  </div>
  <el-form-item :label="t('target')"
    ><el-input
      v-model="model.target"
      :placeholder="model.type === 'http' ? 'https://example.com/health' : 'example.com'"
  /></el-form-item>
  <template v-if="model.type === 'http'"
    ><div class="grid-two">
      <el-form-item label="HTTP Method"
        ><el-select v-model="model.method"
          ><el-option value="GET" /><el-option value="HEAD" /></el-select></el-form-item
      ><el-form-item :label="t('expected')"
        ><el-input-number v-model="model.expected_status" :min="0" :max="599"
      /></el-form-item>
    </div>
    <el-form-item :label="t('certificateDays')"
      ><el-input-number v-model="model.certificate_min_days" :min="0" :max="365" /></el-form-item
    ><el-form-item :label="t('contains')"><el-input v-model="model.contains" /></el-form-item
    ><el-form-item
      ><el-checkbox v-model="model.follow_redirects">{{ t('follow') }}</el-checkbox
      ><el-checkbox v-model="model.insecure_tls">{{ t('insecure') }}</el-checkbox></el-form-item
    ></template
  >
  <el-form-item v-if="['tcp', 'udp'].includes(model.type)" :label="t('port')"
    ><el-input-number v-model="model.port" :min="1" :max="65535"
  /></el-form-item>
  <template v-if="model.type === 'udp'">
    <el-alert :title="t('udpHint')" type="info" :closable="false" />
    <el-form-item :label="t('udpEncoding')"
      ><el-select v-model="model.udp_encoding" placeholder="text"
        ><el-option value="text" label="UTF-8" /><el-option value="hex" label="HEX" /></el-select
    ></el-form-item>
    <el-form-item :label="t('udpPayload')"
      ><el-input v-model="model.udp_payload" type="textarea" :maxlength="4200"
    /></el-form-item>
    <el-form-item :label="t('udpExpected')"
      ><el-input v-model="model.udp_expected" type="textarea" :maxlength="12288"
    /></el-form-item>
  </template>
  <template v-if="model.type === 'dns'"
    ><div class="grid-two">
      <el-form-item :label="t('record')"
        ><el-select v-model="model.dns_record"
          ><el-option
            v-for="kind in ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SOA']"
            :key="kind"
            :value="kind" /></el-select></el-form-item
      ><el-form-item :label="t('resolver')"><el-input v-model="model.resolver" /></el-form-item>
    </div>
    <el-form-item :label="t('dnsExpected')"><el-input v-model="model.expected_dns" /></el-form-item
  ></template>
  <template v-if="['ping', 'mtr'].includes(model.type)"
    ><div class="grid-two">
      <el-form-item :label="t('count')"
        ><el-input-number v-model="model.count" :min="1" :max="20" /></el-form-item
      ><el-form-item :label="t('loss')"
        ><el-input-number v-model="model.max_loss" :min="0" :max="100"
      /></el-form-item>
    </div>
    <el-form-item v-if="model.type === 'ping'" :label="t('size')"
      ><el-input-number v-model="model.size" :min="16" :max="1400" /></el-form-item
  ></template>
  <div class="grid-two">
    <el-form-item :label="t('family')"
      ><el-select v-model="model.family"
        ><el-option
          v-for="family in ['auto', 'ipv4', 'ipv6']"
          :key="family"
          :value="family" /></el-select></el-form-item
    ><el-form-item :label="t('retries')"
      ><el-input-number v-model="model.retries" :min="0" :max="2"
    /></el-form-item>
  </div>
</template>
