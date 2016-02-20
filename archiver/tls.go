package archiver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/nats-io/nats"
	"io/ioutil"
	"strings"
)

type certConfig struct {
	initDone  bool
	certPath  string
	keyPath   string
	caPath    string
	cert      tls.Certificate
	tlsConfig tls.Config
	rootCA    nats.Option
	skipTLS   bool
}

// DefaultCaPath: default location to look for this
// nodes certificate.
var DefaultCertPath string = "certs/my-demo-cert.crt"

// DefaultCaPath: default location to look for the
// this nodes private key.
var DefaultKeyPath string = "private/my-demo-cert.key"

// DefaultCaPath: default location to look for the
// Certificate Authority certificate.
var DefaultCaPath string = "certs/my-demo-ca.pem"

// where to look for certs by default (change this to your installation)
var DefaultCertDir string = "/etc/ssl"

// customize this to your actual cert locations.
func (cc *certConfig) init(tlsDir string) {
	if tlsDir == "" {
		tlsDir = DefaultCertDir
	} else {
		tlsDir = chompSlash(tlsDir) + "/"
	}

	cc.initDone = true

	// these may need customization to your actual default cert paths.
	cc.certPath = tlsDir + DefaultCertPath
	cc.keyPath = tlsDir + DefaultKeyPath
	cc.caPath = tlsDir + DefaultCaPath

	// disable certs for now; this is just for test/demo mode
	cc.skipTLS = true
}

func (cc *certConfig) certLoad() error {
	if !FileExists(cc.certPath) {
		return fmt.Errorf("certLoad: path '%s' does not exist", cc.certPath)
	}
	if !FileExists(cc.keyPath) {
		return fmt.Errorf("certLoad: path '%s' does not exist", cc.keyPath)
	}
	if !FileExists(cc.caPath) {
		return fmt.Errorf("certLoad: path '%s' does not exist", cc.caPath)
	}
	cert, err := tls.LoadX509KeyPair(cc.certPath, cc.keyPath)
	if err != nil {
		return fmt.Errorf("certLoad: error parsing X509 cert='%s'/key='%s', error was: '%v'",
			cc.certPath, cc.keyPath, err)
	}
	cc.cert = cert

	// nats.RootCA will do repeat this, but we detect failure earlier
	// this way and don't bother proceeding down the whole state sequence.
	pool := x509.NewCertPool()
	rootPEM, err := ioutil.ReadFile(cc.caPath)
	if err != nil || rootPEM == nil {
		err = fmt.Errorf("certLoad: error loading "+
			"rootCA file '%s': %v", cc.caPath, err)
		return err
	}
	ok := pool.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return fmt.Errorf("certLoad: failed to parse root certificate from %q", cc.caPath)
	}

	cc.rootCA = nats.RootCAs(cc.caPath)

	cc.tlsConfig = tls.Config{
		Certificates: []tls.Certificate{cc.cert},
		MinVersion:   tls.VersionTLS12,
	}
	return nil
}

func chompSlash(path string) string {
	if strings.HasSuffix(path, "/") {
		return path[:len(path)-1]
	}
	return path
}
