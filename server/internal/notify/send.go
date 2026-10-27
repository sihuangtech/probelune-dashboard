// Package notify 发送最小化通知，不在消息中包含探测凭据或响应正文。
package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"time"
)

type Message struct {
	InstanceID  string   `json:"instance_id"`
	ID          string   `json:"event_id"`
	Instance    string   `json:"instance"`
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	At          int64    `json:"at"`
	URL         string   `json:"url"`
	InstanceURL string   `json:"instance_url"`
	Affected    []string `json:"affected_nodes,omitempty"`
}

func Send(ctx context.Context, ch model.Channel, m Message) error {
	switch ch.Type {
	case "webhook":
		return webhook(ctx, ch, m)
	case "smtp":
		return email(ctx, ch, m)
	default:
		return fmt.Errorf("unsupported channel")
	}
}
func webhook(ctx context.Context, ch model.Channel, m Message) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	stamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(ch.Secret))
	mac.Write([]byte(stamp + "."))
	mac.Write(body)
	req, err := http.NewRequestWithContext(ctx, "POST", ch.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Probelune-Timestamp", stamp)
	req.Header.Set("X-Probelune-Signature", hex.EncodeToString(mac.Sum(nil)))
	transport := &http.Transport{DialContext: ch.Policy.Dial("auto"), TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	client := http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Webhook 连接失败")
	}
	defer res.Body.Close()
	io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Webhook HTTP %d", res.StatusCode)
	}
	return nil
}
func email(ctx context.Context, ch model.Channel, m Message) error {
	address := net.JoinHostPort(ch.SMTPHost, strconv.Itoa(ch.SMTPPort))
	conn, err := ch.Policy.Dial("auto")(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("SMTP 连接失败")
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	tlsConfig := &tls.Config{ServerName: ch.SMTPHost, MinVersion: tls.VersionTLS12}
	if ch.TLSMode == "tls" {
		secure := tls.Client(conn, tlsConfig)
		if err = secure.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("SMTP TLS 校验失败")
		}
		conn = secure
	}
	client, err := smtp.NewClient(conn, ch.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()
	if ch.TLSMode == "starttls" {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("SMTP STARTTLS 失败")
		}
	}
	if ch.Username != "" {
		if err = client.Auth(smtp.PlainAuth("", ch.Username, ch.Password, ch.SMTPHost)); err != nil {
			return fmt.Errorf("SMTP 身份验证失败")
		}
	}
	from, err := mail.ParseAddress(ch.From)
	if err != nil {
		return err
	}
	to, err := mail.ParseAddressList(ch.To)
	if err != nil {
		return err
	}
	if err = client.Mail(from.Address); err != nil {
		return err
	}
	for _, recipient := range to {
		if err = client.Rcpt(recipient.Address); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(m, "", "  ")
	_, err = fmt.Fprintf(writer, "From: %s\r\nTo: %s\r\nSubject: Probelune: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", ch.From, ch.To, m.Kind, raw)
	if err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
