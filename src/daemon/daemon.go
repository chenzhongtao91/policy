package daemon

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"api"
	"meta"
	"store/etcd"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

type daemon struct {
	Router     *mux.Router
	GlobalLock *sync.RWMutex
	PendingOps *metadata.PendingSet
	daemonConfig
}

const (
	CFG_POSTFIX = ".json"
	CONFIGFILE  = "policy.cfg"
	LOCKFILE    = "lock"
)

var (
	lockFile *os.File
	logFile  *os.File

	log = logrus.WithFields(logrus.Fields{"pkg": "daemon"})
)

type daemonConfig struct {
	Root       string
	HostList   []string
	MasterNode string
}

func (c *daemonConfig) ConfigFile() (string, error) {
	if c.Root == "" {
		return "", fmt.Errorf("BUG: Invalid empty daemon config path")
	}
	return filepath.Join(c.Root, CONFIGFILE), nil
}

func (s *daemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	info := fmt.Sprintf("Handler not found: %v %v", r.Method, r.RequestURI)
	log.Errorf(info)
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(info))
}

type requestHandler func(version string, w http.ResponseWriter, r *http.Request, objs map[string]string) error

func makeHandlerFunc(method string, route string, version string, f requestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("User-Agent"), "Policy-Client/") {
			userAgent := strings.Split(r.Header.Get("User-Agent"), "/")
			if len(userAgent) == 2 && userAgent[1] != version {
				http.Error(w, fmt.Errorf("client version %v doesn't match with server %v", userAgent[1], version).Error(), http.StatusNotFound)
				return
			}
		}
		if err := f(version, w, r, mux.Vars(r)); err != nil {
			log.Errorf("Handler for %s %s returned error: %s", method, route, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func createRouter(s *daemon) *mux.Router {
	router := mux.NewRouter()
	m := map[string]map[string]requestHandler{
		"GET": {
			"/volume/":     s.doVolumeGet,
			"/volume/list": s.doVolumeList,
			"/host/":       s.doHostGet,
			"/host/list":   s.doHostList,
			"/device/":     s.doDeviceGet,
			"/device/list": s.doDeviceList,
		},
		"POST": {
			"/volume/create": s.doVolumeCreate,
			"/volume/attach": s.doVolumeAttach,
			"/volume/detach": s.doVolumeDetach,
			"/host/add":      s.doHostAdd,
			"/device/add":    s.doDeviceAdd,
		},
		"DELETE": {
			"/volume/": s.doVolumeDelete,
			"/host/":   s.doHostDel,
			"/device/": s.doDeviceDel,
		},
	}

	for method, routes := range m {
		for route, f := range routes {
			log.Debugf("Registering %s, %s", method, route)
			handler := makeHandlerFunc(method, route, api.API_VERSION, f)
			router.Path("/v{version:[0-9.]+}" + route).Methods(method).HandlerFunc(handler)
			router.Path(route).Methods(method).HandlerFunc(handler)
		}
	}
	router.NotFoundHandler = s

	return router
}

func daemonEnvironmentSetup(c *cli.Context) error {
	var err error

	root := c.String("root")
	if root == "" {
		return fmt.Errorf("Have to specific root directory")
	}
	if err := MkdirIfNotExists(root); err != nil {
		return fmt.Errorf("Invalid root directory:", err)
	}

	lockPath := filepath.Join(root, LOCKFILE)
	if lockFile, err = LockFile(lockPath); err != nil {
		return fmt.Errorf("Failed to lock the file at %v: %v", lockPath, err.Error())
	}

	logrus.SetLevel(logrus.DebugLevel)
	logName := c.String("log")
	if logName != "" {
		logFile, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(logFile)
	} else {
		logrus.SetOutput(os.Stdout)
	}

	return nil
}

func environmentCleanup() {
	log.Debug("Cleaning up environment...")
	if lockFile != nil {
		UnlockFile(lockFile)
	}
	if logFile != nil {
		logFile.Close()
	}
	if r := recover(); r != nil {
		api.ResponseLogAndError(r)
		os.Exit(1)
	}
}

func daemonMetadataSetup(s *daemon) error {
	etcd.NewStore()

	s.PendingOps = metadata.PendingSetSetup()

	go s.PendingOps.MetadataUpdater()

	return nil
}

// Start the daemon
func Start(sockFile string, c *cli.Context) error {
	var err error

	if err = daemonEnvironmentSetup(c); err != nil {
		return err
	}
	defer environmentCleanup()

	root := c.String("root")
	s := &daemon{}
	config := &daemonConfig{
		Root: root,
	}
	exists, err := ObjectExists(config)
	if err != nil {
		return err
	}
	if exists {
		log.Debug("Found existing config. Ignoring command line opts, loading config from ", root)
		if err := ObjectLoad(config); err != nil {
			return nil
		}
	} else {
		hostList := c.StringSlice("hosts")
		if len(hostList) == 0 {
			return fmt.Errorf("Missing or invalid parameters")
		}
		log.Debug("Creating config at ", root)

		config.HostList = hostList
		config.MasterNode = hostList[0]
	}

	s.daemonConfig = *config

	s.GlobalLock = &sync.RWMutex{}

	if err := ObjectSave(config); err != nil {
		return err
	}

	s.Router = createRouter(s)

	if err := MkdirIfNotExists(filepath.Dir(sockFile)); err != nil {
		return err
	}

	//This should be safe because lock file prevent starting daemon twice
	if _, err := os.Stat(sockFile); err == nil {
		log.Warnf("Remove previous sockfile at %v", sockFile)
		if err := os.Remove(sockFile); err != nil {
			return err
		}
	}

	l, err := net.Listen("unix", sockFile)
	if err != nil {
		fmt.Println("listen err", err)
		return err
	}
	defer l.Close()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	log.Debug("aaaaaaaaaaa")
	daemonMetadataSetup(s)
	log.Debug("bbbbbbbbbb")

	ln, err := net.Listen("tcp", ":9876")
	if err != nil {
		fmt.Println("listen err", err)
		return err
	}

	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Printf("Caught signal %s: shutting down.\n", sig)
		done <- true
	}()

	go func() {
		err = http.Serve(l, s.Router)
		if err != nil {
			log.Error("http server error", err.Error())
		}
		done <- true
	}()

	go func() {
		err = http.Serve(ln, s.Router)
		if err != nil {
			log.Error("http server error", err.Error())
		}
		done <- true
	}()

	<-done
	return nil
}
