package email

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestSMTPSenderGivesUpOnAnUnresponsiveServer(t *testing.T) {
	silentListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start silent server: %v", err)
	}
	defer silentListener.Close()
	go func() {
		for {
			connection, err := silentListener.Accept()
			if err != nil {
				return
			}
			defer connection.Close()
		}
	}()

	_, port, _ := net.SplitHostPort(silentListener.Addr().String())
	portNumber, _ := strconv.Atoi(port)
	sender := NewSMTPSender(SMTPConfig{
		Host:    "127.0.0.1",
		Port:    portNumber,
		From:    Address{Email: "no-reply@nebula.test"},
		Timeout: 200 * time.Millisecond,
	})

	sendStartedAt := time.Now()
	err = sender.Send(context.Background(), Message{To: Address{Email: "pilot@nebula.test"}, Subject: "Hi", TextBody: "hi", HTMLBody: "hi"})

	if err == nil {
		t.Fatal("expected an error from a server that never answers")
	}
	if elapsed := time.Since(sendStartedAt); elapsed > 2*time.Second {
		t.Fatalf("send took %v, the timeout must stop it", elapsed)
	}
}
