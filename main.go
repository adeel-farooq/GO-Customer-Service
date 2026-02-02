package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go-cloud-customer/constants"
	"go-cloud-customer/db"
	"go-cloud-customer/middleware"
	"go-cloud-customer/services/registerMethods"
	"log"
	"net/rpc"
	"os"
)

func main() {
	// Configuration loading
	baseDir, _ := os.Getwd()
	_, err := constants.InitializeEnvironmentVariables(baseDir)
	if err != nil {
		log.Fatal("failed to load configuration", err)
	}

	err = db.InitDB()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	/* Create RPC server */
	rpcServer := rpc.NewServer()

	/* Register services on the RPC server */
	registerMethods.RegisterCustomerServices(rpcServer)

	certFile := fmt.Sprintf("ca-%s.crt", constants.Cfg.Environment)
	keyFile := fmt.Sprintf("ca-%s.key", constants.Cfg.Environment)

	/* Load TLS certificates */
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load server certificates: %v", err)
	}

	/* Load CA certificate */
	caCert, err := os.ReadFile(certFile) /* Your CA certificate */
	if err != nil {
		log.Fatalf("Failed to load CA certificate: %v", err)
	}

	/* Create a CA certificate pool for validating client certificates */
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to append CA certificate to pool")
	}

	/* Configure TLS */
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,                     /* Use CA pool to verify client certificates */
		ClientAuth:   tls.RequireAndVerifyClientCert, /* Enforce client certificate validation */
	}

	/* Start listening on a secure port */
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%s", constants.Cfg.PORT), config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Customer microservice running with TLS on port %s", constants.Cfg.PORT)

	/* Serve connections */
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		go middleware.ServeConnWithRecovery(rpcServer, conn)
	}
}
