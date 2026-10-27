# Notifications / 通知配置

Create channels in Alerts, test each, then select channel IDs on monitors. Node offline/online events use enabled channels. Target outage and recovery each use a stable event/kind/channel delivery key. Delivery logs retain attempt count, status and errors; failure retries after 10 and 20 seconds, with three attempts total. Maintenance and timed silence suppress notifications, not observations.

## Webhook

Use HTTPS and set a random signing secret. A generated secret is shown on save; store it on the receiver. Explicit private CIDRs are required for internal receivers. Redirects are rejected. The receiver should return a 2xx status quickly, verify freshness and deduplicate by event/kind.

Signature: lower-case hex HMAC-SHA256(secret, `X-Probelune-Timestamp + "." + exactRawBody`). Compare in constant time and reject timestamps outside your chosen replay window (for example five minutes). Parse JSON **after** verifying the original bytes. `event_id`, `kind`, `instance`, monitor name, affected node IDs and an authenticated event URL are sent; probe authentication headers/response bodies are not. Test messages have no incident ID.

## SMTP

Configure host, port, From, To, optional username/password and TLS mode (`tls`, normally 465; `starttls`, normally 587). Plaintext SMTP and automatic TLS downgrade are not supported. Trust your private CA through the operating system if needed. Sender authorization, delivery quotas and spam filtering belong to your provider; channel acceptance is not inbox delivery proof.

## Optional native push

This release connects directly to **your own Firebase project**. No Probelune-operated gateway or universal app credentials are embedded. Build the mobile client with that project's public Firebase configuration and enable the FCM API. On the server, mount a Firebase service-account JSON outside the repository/data backup and set `GOOGLE_APPLICATION_CREDENTIALS` to its path. On iOS, configure APNs credentials and Push Notifications capability for the app identifier in the same Firebase project.

Create an enabled `push` channel, select it on monitors, then opt into notifications in the app on each instance. Device registration is user-scoped and revocable from Settings; disabled users are skipped. Push messages carry only event ID, instance URL and event kind plus generic visible text. The app fetches details from the private dashboard; use VPN when that dashboard is only privately reachable.

当前 Firebase Messaging 17 使用 `register()` / `onRegistered` 返回 installation ID，服务端 HTTP v1 使用 `message.fid`。配置与签名步骤见[移动端文档](https://github.com/sihuangtech/probelune-mobile/blob/main/docs/RELEASE.md)。没有 Firebase/APNs 凭据或设备的环境只能验证未配置状态，不能证明原生推送实际送达。没有 Google Play 服务或无法访问 FCM 的 Android 环境不承诺原生推送，可使用 SMTP/Webhook 与手动刷新。

References: [Firebase Flutter setup](https://firebase.google.com/docs/cloud-messaging/flutter/get-started), [FCM HTTP v1 message resource](https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages).

Channels may configure `reminder_minutes` (0 by default, up to 1440). Only a still-confirmed active failure is reminded; stable time slots deduplicate repeated worker passes. Push also carries the stable instance ID so a saved LAN URL can match the same public-facing dashboard.
