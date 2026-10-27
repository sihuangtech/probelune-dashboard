package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"os"
)

// Push 使用部署者自己的 FCM HTTP v1 凭据；仅推送事件标识，详情从私有面板读取。
func Push(ctx context.Context, token string, m Message) error {
	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if path == "" {
		return fmt.Errorf("未配置 FCM 服务账号文件")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("无法读取 FCM 凭据")
	}
	creds, err := google.CredentialsFromJSON(ctx, raw, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return fmt.Errorf("FCM 凭据格式不正确")
	}
	var info struct {
		ProjectID string `json:"project_id"`
	}
	if json.Unmarshal(raw, &info) != nil || info.ProjectID == "" {
		return fmt.Errorf("缺少 Firebase project_id")
	}
	access, err := creds.TokenSource.Token()
	if err != nil {
		return fmt.Errorf("FCM 授权失败")
	}
	payload := map[string]any{"message": map[string]any{"fid": token, "notification": map[string]string{"title": "巡汐 Probelune", "body": "监测状态发生变化，点击查看"}, "data": map[string]string{"event_id": m.ID, "instance_url": m.InstanceURL, "instance_id": m.InstanceID, "kind": m.Kind}, "android": map[string]string{"priority": "high"}, "apns": map[string]any{"headers": map[string]string{"apns-priority": "10"}, "payload": map[string]any{"aps": map[string]string{"sound": "default"}}}}}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/v1/projects/"+info.ProjectID+"/messages:send", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+access.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("FCM 请求失败")
	}
	defer res.Body.Close()
	io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("FCM HTTP %d", res.StatusCode)
	}
	return nil
}
