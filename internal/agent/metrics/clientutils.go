package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
)

func CalculateHash(data []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}
