package main

import (
	"net"
	"github.com/op/go-logging"
	"github.com/DiegoAscanio/smtpd"
	"strings"
	b64 "encoding/base64"
)

var log = logging.MustGetLogger("smtp trigger")
var format = logging.MustStringFormatter(
                `%{color}%{time:15:04:05.000} %{shortfunc} | ${level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func sockWrite(addr, data string) {
                typ := "unix"
                conn, err := net.DialUnix(typ, nil, &net.UnixAddr{addr, typ})
                if err != nil {
                                log.Errorf("open socket failed: %e", err)
                                return
                }
                defer conn.Close()

                _, err = conn.Write([]byte(data))
                if err != nil {
                                log.Errorf("write data failed: %e", err)
                                return
                }
}

func handler(peer smtpd.Peer, env smtpd.Envelope) error {
	if len(env.Recipients) != 1 {
		log.Errorf("Too Many recipients from %s", env.Sender)
		return smtpd.Error{Code: 452, Message: "Too many recipients"}
	}
	if env.Recipients[0] != "zm@trigger.smtp" {
		return smtpd.Error{Code: 550, Message: "Invalid recipient"}
	}
	// processa o assunto da mensagem oriunda da camera para enviar a mensagem ao zm
	encMessage := strings.Split(string(env.Data), "\n")[3]
	encMessage = strings.Replace(encMessage, "Subject: =?UTF-8?B?", "", 1)
	decMessage, _ := b64.StdEncoding.DecodeString(encMessage)
	zmMessage := string(decMessage)
	// fim do processamento

	// escreve mensagem no socket
	log.Infof("\"%s\" from %s", zmMessage, env.Sender)
	sockWrite("/var/run/zm/zmtrigger.sock", zmMessage)
	return nil
}

func authenticator(peer smtpd.Peer, username string, password string) error {
	if username != "zm" && password != "zm" {
		return smtpd.Error{Code: 550, Message: "Denied"}
	} else {
		return nil
	}
}

func main() {
	server := &smtpd.Server {
		Handler: handler,
		Authenticator: authenticator,
	}
	server.ListenAndServe("0.0.0.0:2525")
}
