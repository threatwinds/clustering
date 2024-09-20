package helpers

import (
	"fmt"
	"net"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/threatwinds/logger"
)

func GetMainIP() (string, *logger.Error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", go_sdk.Logger().ErrorF(err.Error())
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func TestPort(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), 30*time.Second)
	if err != nil {
		return false
	}

	if conn != nil {
		defer conn.Close()
		return true
	}

	return false
}
