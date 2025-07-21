package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
)

func GenerateTLSConfig() *tls.Config {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"QUIC Server"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	keyPEM := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}
	return &tls.Config{Certificates: []tls.Certificate{keyPEM}, NextProtos: []string{"data-transfer"}}
}

const addr = ":4242"

func main() {
	dirPtr := flag.String("dir", ".", "Directory to serve files from")
	flag.Parse()

	listener, err := quic.ListenAddr(addr, GenerateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server listening on %s, serving from %s", addr, *dirPtr)

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn, *dirPtr)
	}
}

func handleConnection(conn *quic.Conn, baseDir string) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		log.Println("Stream error:", err)
		return
	}
	defer stream.Close()

	reader := bufio.NewReader(stream)
	cmdLine, _ := reader.ReadString('\n')
	cmdLine = strings.TrimSpace(cmdLine)

	if cmdLine == "ls" {
		files, _ := os.ReadDir(baseDir)
		for _, file := range files {
			if !file.IsDir() {
				fmt.Fprintln(stream, file.Name())
			}
		}
	} else if strings.HasPrefix(cmdLine, "get ") {
		filename := strings.TrimPrefix(cmdLine, "get ")
		fullpath := filepath.Join(baseDir, filename)

		file, err := os.Open(fullpath)
		if err != nil {
			fmt.Fprintf(stream, "ERR: %v\n", err)
			return
		}
		defer file.Close()

		io.Copy(stream, file)
	}
}
