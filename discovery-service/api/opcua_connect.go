package api

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type Session struct {
	Client       *opcua.Client
	LastAccessed time.Time
}

// Global map to store active connections
var (
	sessions = make(map[string]*Session)
	mu       sync.Mutex
)

type CreateConnectionInput struct {
	Body CreateConnectionBody
}

type CreateConnectionBody struct {
	Host           string         `json:"host" example:"servername.com" doc:"The host name of the opc ua server" required:"true"`
	Port           uint16         `json:"port" example:"4840" doc:"The port of the opc ua server" required:"true"`
	Policy         string         `json:"policy" example:"Basic256" doc:"The policy of the opc ua server" required:"true"`
	Mode           string         `json:"mode" example:"SignAndEncrypt" doc:"The encryption mode of the opc ua server" required:"true"`
	Authentication string         `json:"authentication" example:"User&Password" doc:"The encryption mode of the opc ua server" enum:"User&Password,Anonymous"`
	OPCCredentials OPCCredentials `json:"credentials" required:"false"`
	OPCCertificate OPCCertificate `json:"certificate" required:"false"`
}

type OPCCredentials struct {
	Username string `json:"username" example:"User" doc:"Username, if Authentication is set to 'Username and Password'" `
	Password string `json:"password" example:"User" doc:"Password, if Authentication is set to 'Username and Password'" `
}

type OPCCertificate struct {
	Certificate string `json:"certificate" doc:"PEM Encoded certificate, if policy is not 'None' " `
	PrivateKey  string `json:"privateKey"  doc:"PEM Encoded private key, if policy is not 'None' " `
}

type CreateConnectionOutput struct {
	Body struct {
		UUID string `json:"uuid"`
	}
}

func HouseKeeper() { // Async Func for housekeeping

	for {

		var toClose []*opcua.Client

		mu.Lock()
		for key, sess := range sessions {
			if time.Since(sess.LastAccessed) > 10*time.Minute {
				if sess.Client != nil {
					toClose = append(toClose, sess.Client)
				}
				delete(sessions, key)
			}
		}
		mu.Unlock()

		for _, c := range toClose {
			// Context with a timeout ensures a hanging close doesn't block the loop
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			c.Close(ctx)
			cancel()
		}

		time.Sleep(1 * time.Minute)
	}
}

func (a *AppState) CreateConnection(ctx context.Context, input *CreateConnectionInput) (*CreateConnectionOutput, error) {

	client, err := input.Body.CreateClient(ctx)

	if err != nil {
		return nil, err
	}

	sessionID := uuid.New().String()

	mu.Lock()
	defer mu.Unlock()

	sessions[sessionID] = &Session{
		Client:       client,
		LastAccessed: time.Now(),
	}

	resp := &CreateConnectionOutput{}
	resp.Body.UUID = sessionID

	return resp, nil

}

func (c *CreateConnectionBody) CreateClient(ctx context.Context) (*opcua.Client, error) {

	con_string := fmt.Sprintf("opc.tcp://%s:%d", c.Host, c.Port)

	eps, err := opcua.GetEndpoints(ctx, con_string)

	if err != nil {
		return nil, err
	}

	if len(eps) < 1 {
		return nil, fmt.Errorf("no endpoints found - check configuration")
	}

	ep, err := opcua.SelectEndpoint(eps, c.Policy, ua.MessageSecurityModeFromString(c.Mode))

	if err != nil {
		return nil, err
	}

	opts := []opcua.Option{
		opcua.ApplicationName("geist-api"),
		opcua.AutoReconnect(true),
		opcua.ReconnectInterval(10 * time.Second),
		opcua.SecurityPolicy(c.Policy),
		opcua.SecurityMode(ua.MessageSecurityModeFromString(c.Mode)),
	}

	switch c.Authentication {

	case "User&Password":
		opts = append(opts, opcua.AuthUsername(c.OPCCredentials.Username, c.OPCCredentials.Password))
		opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeUserName))

	case "Certificate": // Cert Config goes here, but has to be evaluated first

	default:
		opts = append(opts, opcua.AuthAnonymous())
		opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous))
	}

	if c.Policy != "None" {
		block_cert, _ := pem.Decode([]byte(c.OPCCertificate.Certificate))
		if block_cert == nil || block_cert.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("failed to decode PEM block: check if certificate is valid PEM")
		}

		block_key, _ := pem.Decode([]byte(c.OPCCertificate.PrivateKey))
		if block_key == nil {
			return nil, fmt.Errorf("failed to decode PEM block: check if private key is valid PEM")
		}

		var key interface{}
		var err error
		if key, err = x509.ParsePKCS1PrivateKey(block_key.Bytes); err != nil {
			if key, err = x509.ParsePKCS8PrivateKey(block_key.Bytes); err != nil {
				return nil, fmt.Errorf("failed to parse private key: %v", err)
			}
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parsed key is not an RSA private key (unsupported algorithm)")
		}

		opts = append(opts, opcua.Certificate(block_cert.Bytes))
		opts = append(opts, opcua.PrivateKey(rsaKey))
	}

	client, err := opcua.NewClient(con_string, opts...)

	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil

}
