package etcd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"
)

var (
	ErrNoAvailSrc = errors.New("no available argument and stdin")

	// the maximum amount of time a dial will wait for a connection to setup.
	// 30s is long enough for most of the network conditions.
	defaultDialTimeout = 30 * time.Second
)

// trimsplit slices s into all substrings separated by sep and returns a
// slice of the substrings between the separator with all leading and trailing
// white space removed, as defined by Unicode.
func trimsplit(s, sep string) []string {
	raw := strings.Split(s, ",")
	trimmed := make([]string, 0)
	for _, r := range raw {
		trimmed = append(trimmed, strings.TrimSpace(r))
	}
	return trimmed
}

func argOrStdin(args []string, stdin io.Reader, i int) (string, error) {
	if i < len(args) {
		return args[i], nil
	}
	bytes, err := ioutil.ReadAll(stdin)
	if string(bytes) == "" || err != nil {
		return "", ErrNoAvailSrc
	}
	return string(bytes), nil
}

func getPeersFlagValue() []string {

	peerstr := "http://127.0.0.1:2379,http://127.0.0.1:4001"

	return strings.Split(peerstr, ",")
}

func getDomainDiscoveryFlagValue() ([]string, error) {
	return []string{}, nil

}

func getEndpoints() ([]string, error) {
	eps, err := getDomainDiscoveryFlagValue()
	if err != nil {
		return nil, err
	}

	// If domain discovery returns no endpoints, check peer flag
	if len(eps) == 0 {
		eps = getPeersFlagValue()
	}

	for i, ep := range eps {
		u, err := url.Parse(ep)
		if err != nil {
			return nil, err
		}

		if u.Scheme == "" {
			u.Scheme = "http"
		}

		eps[i] = u.String()
	}

	return eps, nil
}

func getTransport() (*http.Transport, error) {
	// Use an environment variable if nothing was supplied on the
	// command line

	cafile := os.Getenv("ETCDCTL_CA_FILE")
	certfile := os.Getenv("ETCDCTL_CERT_FILE")
	keyfile := os.Getenv("ETCDCTL_KEY_FILE")

	tls := transport.TLSInfo{
		CAFile:   cafile,
		CertFile: certfile,
		KeyFile:  keyfile,
	}
	return transport.NewTransport(tls, defaultDialTimeout)
}

func getUsernamePasswordFromFlag(usernameFlag string) (username string, password string, err error) {
	return getUsernamePassword("Password: ", usernameFlag)
}

func getUsernamePassword(prompt, usernameFlag string) (username string, password string, err error) {
	colon := strings.Index(usernameFlag, ":")
	if colon == -1 {
		username = usernameFlag
		// Prompt for the password.
		password, err = speakeasy.Ask(prompt)
		if err != nil {
			return "", "", err
		}
	} else {
		username = usernameFlag[:colon]
		password = usernameFlag[colon+1:]
	}
	return username, password, nil
}

func mustNewKeyAPI() client.KeysAPI {
	return client.NewKeysAPI(mustNewClient())
}

func mustNewMembersAPI(c *cli.Context) client.MembersAPI {
	return client.NewMembersAPI(mustNewClient())
}

func mustNewClient() client.Client {
	hc, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	debug := false

	if debug {
		client.EnablecURLDebug()
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.DefaultRequestTimeout)
	err = hc.Sync(ctx)
	cancel()
	if err != nil {
		if err == client.ErrNoEndpoints {
			fmt.Fprintf(os.Stderr, "etcd cluster has no published client endpoints.\n")
			fmt.Fprintf(os.Stderr, "Try '--no-sync' if you want to access non-published client endpoints(%s).\n", strings.Join(hc.Endpoints(), ","))
			return nil
		}

		if isConnectionError(err) {
			return nil
		}

		// fail-back to try sync cluster with peer API. this is for making etcdctl work with etcd 0.4.x.
		// TODO: remove this when we deprecate the support for etcd 0.4.
		eps, serr := syncWithPeerAPI(ctx, hc.Endpoints())
		if serr != nil {
			if isConnectionError(serr) {
				return nil
			} else {
				return nil
			}
		}
		err = hc.SetEndpoints(eps)
		if err != nil {
			return nil
		}
	}
	if debug {
		fmt.Fprintf(os.Stderr, "got endpoints(%s) after sync\n", strings.Join(hc.Endpoints(), ","))
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Cluster-Endpoints: %s\n", strings.Join(hc.Endpoints(), ", "))
	}

	return hc
}

func isConnectionError(err error) bool {
	switch t := err.(type) {
	case *client.ClusterError:
		for _, cerr := range t.Errors {
			if !isConnectionError(cerr) {
				return false
			}
		}
		return true
	case *net.OpError:
		if t.Op == "dial" || t.Op == "read" {
			return true
		}
		return isConnectionError(t.Err)
	case net.Error:
		if t.Timeout() {
			return true
		}
	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return true
		}
	}
	return false
}

func mustNewClientNoSync() client.Client {
	hc, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return hc
}

func newClient() (client.Client, error) {
	eps, err := getEndpoints()
	if err != nil {
		return nil, err
	}

	tr, err := getTransport()
	if err != nil {
		return nil, err
	}

	cfg := client.Config{
		Transport:               tr,
		Endpoints:               eps,
		HeaderTimeoutPerRequest: time.Second,
	}

	return client.New(cfg)
}

func contextWithTotalTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// syncWithPeerAPI syncs cluster with peer API defined at
// https://github.com/coreos/etcd/blob/v0.4.9/server/server.go#L311.
// This exists for backward compatibility with etcd 0.4.x.
func syncWithPeerAPI(ctx context.Context, knownPeers []string) ([]string, error) {
	tr, err := getTransport()
	if err != nil {
		return nil, err
	}

	var (
		body []byte
		resp *http.Response
	)
	for _, p := range knownPeers {
		var req *http.Request
		req, err = http.NewRequest("GET", p+"/v2/peers", nil)
		if err != nil {
			continue
		}
		resp, err = tr.RoundTrip(req)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Parse the peers API format: https://github.com/coreos/etcd/blob/v0.4.9/server/server.go#L311
	return strings.Split(string(body), ", "), nil
}
