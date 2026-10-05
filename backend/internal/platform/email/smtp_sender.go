package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

const DefaultSMTPTimeout = 10 * time.Second

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     Address
	Timeout  time.Duration
}

type SMTPSender struct {
	config SMTPConfig
}

func NewSMTPSender(config SMTPConfig) *SMTPSender {
	if config.Timeout == 0 {
		config.Timeout = DefaultSMTPTimeout
	}
	return &SMTPSender{config: config}
}

func (sender *SMTPSender) Send(ctx context.Context, message Message) error {
	encodedMessage, err := buildMIMEMessage(sender.config.From, message)
	if err != nil {
		return err
	}

	sendContext, cancelSend := context.WithTimeout(ctx, sender.config.Timeout)
	defer cancelSend()

	serverAddress := net.JoinHostPort(sender.config.Host, strconv.Itoa(sender.config.Port))
	var dialer net.Dialer
	connection, err := dialer.DialContext(sendContext, "tcp", serverAddress)
	if err != nil {
		return fmt.Errorf("connect to smtp server %s: %w", serverAddress, err)
	}
	if deadline, hasDeadline := sendContext.Deadline(); hasDeadline {
		_ = connection.SetDeadline(deadline)
	}

	smtpClient, err := smtp.NewClient(connection, sender.config.Host)
	if err != nil {
		_ = connection.Close()
		return fmt.Errorf("start smtp session: %w", err)
	}
	defer smtpClient.Close()

	if err := sender.deliver(smtpClient, message.To.Email, encodedMessage); err != nil {
		return fmt.Errorf("send email to %s: %w", message.To.Email, err)
	}
	return nil
}

func (sender *SMTPSender) deliver(smtpClient *smtp.Client, recipientEmail string, encodedMessage []byte) error {
	if supportsStartTLS, _ := smtpClient.Extension("STARTTLS"); supportsStartTLS {
		if err := smtpClient.StartTLS(&tls.Config{ServerName: sender.config.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if sender.config.Username != "" {
		authentication := smtp.PlainAuth("", sender.config.Username, sender.config.Password, sender.config.Host)
		if err := smtpClient.Auth(authentication); err != nil {
			return err
		}
	}
	if err := smtpClient.Mail(sender.config.From.Email); err != nil {
		return err
	}
	if err := smtpClient.Rcpt(recipientEmail); err != nil {
		return err
	}
	dataWriter, err := smtpClient.Data()
	if err != nil {
		return err
	}
	if _, err := dataWriter.Write(encodedMessage); err != nil {
		return err
	}
	if err := dataWriter.Close(); err != nil {
		return err
	}
	return smtpClient.Quit()
}

func formatAddress(address Address) string {
	return (&mail.Address{Name: address.Name, Address: address.Email}).String()
}

func buildMIMEMessage(from Address, message Message) ([]byte, error) {
	var messageBuffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&messageBuffer)

	messageHeaders := []string{
		"From: " + formatAddress(from),
		"To: " + formatAddress(message.To),
		"Subject: " + mime.QEncoding.Encode("utf-8", message.Subject),
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
		"Message-ID: " + newMessageID(from.Email),
		"MIME-Version: 1.0",
		"Content-Type: multipart/alternative; boundary=" + bodyWriter.Boundary(),
	}
	var headerBuffer bytes.Buffer
	for _, messageHeader := range messageHeaders {
		headerBuffer.WriteString(messageHeader + "\r\n")
	}
	headerBuffer.WriteString("\r\n")

	for _, bodyPart := range []struct {
		contentType string
		content     string
	}{
		{contentType: "text/plain; charset=utf-8", content: message.TextBody},
		{contentType: "text/html; charset=utf-8", content: message.HTMLBody},
	} {
		partWriter, err := bodyWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {bodyPart.contentType},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, err
		}
		encodingWriter := quotedprintable.NewWriter(partWriter)
		if _, err := encodingWriter.Write([]byte(bodyPart.content)); err != nil {
			return nil, err
		}
		if err := encodingWriter.Close(); err != nil {
			return nil, err
		}
	}
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	return append(headerBuffer.Bytes(), messageBuffer.Bytes()...), nil
}

func newMessageID(fromEmail string) string {
	randomBytes := make([]byte, 12)
	_, _ = rand.Read(randomBytes)
	domain := "localhost"
	if atIndex := strings.LastIndex(fromEmail, "@"); atIndex >= 0 {
		domain = fromEmail[atIndex+1:]
	}
	return "<" + hex.EncodeToString(randomBytes) + "@" + domain + ">"
}
