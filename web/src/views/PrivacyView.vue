<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { language } from '../lib/i18n'
const route = useRoute()
const content = {
  en: [
    ['Privacy Policy', 'Effective date: November 4, 2026'],
    [
      'Scope and contact',
      'This policy describes the Probelune mobile application and this Probelune web dashboard. For privacy questions or requests, contact developer@skstudio.cn. Probelune can connect to independently operated, self-hosted dashboards. The operator of the dashboard you choose controls its data and access. Contact that operator for information about their deployment and service providers.',
    ],
    [
      'Information processed',
      'When you connect to a dashboard, the application sends the dashboard address, username and password or single-use pairing code to that dashboard for authentication. The dashboard stores account records and password hashes, issues revocable sessions, and records session expiration and client identification. Monitoring configuration, targets, probe results, diagnostics, incidents, notification destinations and administrative audit records are processed to provide the functions you request. The dashboard server and its network providers may process connection information such as IP addresses in operational or security logs.',
    ],
    [
      'Storage on your device',
      'The mobile app stores dashboard addresses, session credentials and a last-known overview snapshot using operating-system secure storage. The snapshot supports offline viewing for up to 24 hours. Language, appearance and history preferences are stored locally. The web dashboard uses a session cookie for authentication and browser storage for interface preferences.',
    ],
    [
      'Permissions and optional services',
      'Camera access is used when you choose to scan a pairing QR code. Network and local-network access connect to your selected dashboard. Optional notifications require a configured build, a notification service and your permission. When push is enabled, installation identifiers and notification routing information are processed by the dashboard and Firebase/APNs. Push messages can contain an event identifier, instance URL, event type and generic notification text. Those providers process delivery and device metadata under their own policies. The operating system handles content you explicitly export or share.',
    ],
    [
      'Use and disclosure',
      'Information is used for authentication, monitoring, diagnostics, incident handling, notifications and service security. The application contains no bundled advertising or analytics service. Data is sent to the dashboard you select and to services configured by its operator, such as email, webhook or push providers. If public status is enabled, monitors marked public can expose their names and status summaries. Dashboard administrators control users, roles, retention and public visibility.',
    ],
    [
      'Retention and deletion',
      'Dashboard retention is configured by its operator. Default retention is 7 days for raw results and 180 days for events and audit history. Mobile sessions last 30 days unless revoked earlier. Removing an instance clears its local credentials and cache and attempts to revoke its server session and device registration. If the dashboard is unreachable, ask its administrator to revoke them. Operating-system backups and Keychain behavior may retain information independently. To request account or server-record deletion, contact your dashboard operator or developer@skstudio.cn. Retention in backups and required operational records depends on the deployment.',
    ],
    [
      'Security and your choices',
      'Use an HTTPS dashboard address. The application stores credentials through platform secure storage, and the dashboard applies role-based access controls and password hashing. You can revoke sessions, disable notifications, remove saved instances and request access, correction or deletion through your dashboard operator. No transmission or storage mechanism can guarantee absolute security.',
    ],
    [
      'Children and policy changes',
      'Probelune is a technical tool for service administrators and developers and is not directed at children. If you believe a child has provided personal information, contact us. This page will be updated when data practices change, with the effective date shown above.',
    ],
  ],
  zh: [
    ['隐私政策', '生效日期：2026 年 11 月 4 日'],
    [
      '适用范围与联系方式',
      '本政策说明 Probelune 移动应用及本 Probelune 网页面板的数据处理方式。隐私问题或相关请求请联系 developer@skstudio.cn。Probelune 可连接由不同运营者独立部署的面板，你选择的面板运营者控制其数据与访问权限。有关具体部署及服务提供商的信息，请联系该面板运营者。',
    ],
    [
      '处理的信息',
      '连接面板时，应用向该面板提交地址、用户名与密码，或一次性配对码，用于身份验证。面板保存账户记录和密码哈希，签发可撤销的会话，并记录会话到期时间和客户端标识。为提供你请求的功能，面板会处理监测配置、检测目标、探测结果、诊断、故障事件、通知目的地址及管理审计记录。面板服务器及其网络服务提供商可能在运行或安全日志中处理 IP 地址等连接信息。',
    ],
    [
      '设备上的存储',
      '移动应用使用操作系统安全存储保存面板地址、会话凭据和最近一次总览快照。快照用于最长 24 小时的离线查看。语言、外观与历史范围偏好保存在本地。网页面板使用会话 Cookie 完成身份验证，并通过浏览器存储保存界面偏好。',
    ],
    [
      '权限与可选服务',
      '仅在你选择扫描配对二维码时使用摄像头。网络和本地网络权限用于连接所选面板。可选通知需要相应构建配置、通知服务和你的授权。启用推送后，面板与 Firebase/APNs 会处理安装标识及通知路由信息，推送内容可包含事件标识、实例地址、事件类型及通用提示文字。服务商依据各自政策处理投递与设备元数据。你主动导出或分享的内容由操作系统分享功能交付。',
    ],
    [
      '用途与对外提供',
      '信息用于身份验证、监测、诊断、故障处理、通知及服务安全。应用未内置广告或分析统计服务。数据发往你选择的面板，以及面板运营者配置的邮件、Webhook 或推送等服务。开启公开状态页后，被标记为公开的监测可展示名称和状态摘要。面板管理员控制用户、角色、保留期限及公开范围。',
    ],
    [
      '保留与删除',
      '面板数据保留期限由运营者配置，原始结果默认保留 7 天，事件与审计历史默认保留 180 天。移动会话有效期为 30 天，可提前撤销。移除实例会清除本地凭据及缓存，并尝试撤销服务端会话和设备注册。面板不可达时，请联系管理员完成撤销。操作系统备份及钥匙串机制可能独立保留信息。请求删除账户或服务器记录，请联系面板运营者或 developer@skstudio.cn。备份及必要运行记录的保留取决于具体部署。',
    ],
    [
      '安全与选择',
      '请使用 HTTPS 面板地址。应用通过系统安全存储保存凭据，面板使用角色权限控制及密码哈希。你可以撤销会话、关闭通知、移除已保存实例，并向面板运营者请求访问、更正或删除信息。任何传输与存储机制都无法保证绝对安全。',
    ],
    [
      '未成年人及政策更新',
      'Probelune 面向服务管理员和开发者提供技术工具，并非面向儿童设计。如你认为儿童提供了个人信息，请联系我们。数据处理方式发生变化时，本页会更新并标明生效日期。',
    ],
  ],
}
const chinese = computed(
  () =>
    route.path.endsWith('/zh') || (!route.path.endsWith('/en') && language.value.startsWith('zh')),
)
const sections = computed(() => (chinese.value ? content.zh : content.en))
</script>
<template>
  <main class="privacy-page" :lang="chinese ? 'zh-CN' : 'en'">
    <header>
      <router-link to="/">Probelune</router-link>
      <nav aria-label="Language">
        <router-link to="/privacy/zh">简体中文</router-link
        ><router-link to="/privacy/en">English</router-link>
      </nav>
    </header>
    <h1>{{ sections[0]?.[0] }}</h1>
    <p class="muted">{{ sections[0]?.[1] }}</p>
    <section v-for="section in sections.slice(1)" :key="section[0]">
      <h2>{{ section[0] }}</h2>
      <p>{{ section[1] }}</p>
    </section>
    <footer><a href="mailto:developer@skstudio.cn">developer@skstudio.cn</a></footer>
  </main>
</template>
<style scoped>
.privacy-page {
  max-width: 820px;
  margin: 0 auto;
  padding: 28px 24px 56px;
  overflow-wrap: anywhere;
}
header,
nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
}
header {
  padding-bottom: 24px;
  border-bottom: 1px solid var(--line);
  margin-bottom: 36px;
}
header > a {
  font-size: 22px;
  font-weight: 600;
}
h2 {
  font-size: 21px;
  margin: 30px 0 10px;
}
p {
  line-height: 1.9;
}
footer {
  margin-top: 36px;
  border-top: 1px solid var(--line);
  padding-top: 20px;
}
</style>
