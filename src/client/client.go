package client

import (
	//"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	//"net/url"
	"os"
	"time"

	"api"
	"daemon"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

/* define the PolicyClient struct */
type policyClient struct {
	addr      string
	scheme    string
	transport *http.Transport
}

/* define the global vars */
var (
	verboseFlag = "verbose"

	log    = logrus.WithFields(logrus.Fields{"pkg": "client"})
	client policyClient
)

func (c *policyClient) call(method, path string, data interface{}, headers map[string][]string) (io.ReadCloser, int, error) {
	params, err := daemon.EncodeData(data)
	if err != nil {
		return nil, -1, err
	}

	if data != nil {
		if headers == nil {
			headers = make(map[string][]string)
		}
		headers["Context-Type"] = []string{"application/json"}
	}

	body, _, statusCode, err := c.clientRequest(method, path, params, headers)
	return body, statusCode, err
}

func (c *policyClient) httpClient() *http.Client {
	return &http.Client{Transport: c.transport}
}

func getRequestPath(path string) string {
	return fmt.Sprintf("/v1%s", path)
}

func (c *policyClient) clientRequest(method, path string, in io.Reader, headers map[string][]string) (io.ReadCloser, string, int, error) {
	req, err := http.NewRequest(method, getRequestPath(path), in)
	if err != nil {
		return nil, "", -1, err
	}
	req.Header.Set("User-Agent", "Policy-Client/"+api.API_VERSION)
	req.URL.Host = c.addr
	req.URL.Scheme = c.scheme

	resp, err := c.httpClient().Do(req)
	statusCode := -1
	if resp != nil {
		statusCode = resp.StatusCode
	}
	if err != nil {
		return nil, "", statusCode, err
	}
	if statusCode < 200 || statusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, "", statusCode, err
		}
		if len(body) == 0 {
			return nil, "", statusCode, fmt.Errorf("Incompatable version")
		}
		return nil, "", statusCode, fmt.Errorf("Error response from server, %v", string(body))
	}
	return resp.Body, resp.Header.Get("Context-Type"), statusCode, nil
}

func cmdNotFound(c *cli.Context, command string) {
	panic(fmt.Errorf("Unrecognized command: %s", command))
}

// NewCli would generate Policy CLI
func NewCli(version string) *cli.App {
	app := cli.NewApp()
	app.Name = "policy"
	app.Version = version
	app.Author = "Jiang Louis <jiangang.jiang@cloudsoar.com>"
	app.Usage = "A policy manage capable of Blastaar"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "socket, s",
			Value: "/var/run/policy/policy.sock",
			Usage: "Specify unix domain socket for communication between server and client",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Enable debug level log with client or not",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Verbose level output for client",
		},
	}
	app.CommandNotFound = cmdNotFound
	app.Before = initClient
	app.Commands = []cli.Command{
		daemonCmd,
		infoCmd,
		VolumeCmds,
		DeviceCmds,
		HostCmds,
	}
	return app
}

// connect to the specific server (unix sock or ip:port)
func initClient(c *cli.Context) error {
	sockFile := c.GlobalString("socket")
	if sockFile == "" {
		return fmt.Errorf("Require unix domain socket location")
	}
	logrus.SetOutput(os.Stderr)
	debug := c.GlobalBool("debug")
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	client.addr = sockFile
	client.scheme = "http"
	client.transport = &http.Transport{
		DisableCompression: true,
		Dial: func(_, _ string) (net.Conn, error) {
			return net.DialTimeout("unix", sockFile, 10*time.Second)
		},
	}
	return nil
}

func sendRequest(method, request string, data interface{}) (io.ReadCloser, error) {
	log.Debugf("Sending request %v %v", method, request)
	if data != nil {
		log.Debugf("With data %+v", data)
	}
	rc, _, err := client.call(method, request, data, nil)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func sendRequestAndPrint(method, request string, data interface{}) error {
	rc, err := sendRequest(method, request, data)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return nil
	}
	defer rc.Close()

	b, err := ioutil.ReadAll(rc) // Get the response from the Http Server
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return nil
	}
	//log.Debugf("%+v", b)
	fmt.Println(string(b)) // Just print the response ([]byte)
	return nil
}

func PrintErrorInfo(err error) {
	fmt.Println("Error: ", err.Error())
}
