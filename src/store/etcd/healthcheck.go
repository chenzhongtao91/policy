// healthcheck
package etcd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func (estore *EtcdStoreDriver) HealthCheck() (r string, err error) {
	tr, err := getTransport()
	if err != nil {
		return "", err
	}

	hc := http.Client{
		Transport: tr,
	}

	cln := mustNewClientNoSync()
	mi := client.NewMembersAPI(cln)
	ms, err := mi.List(context.TODO())
	if err != nil {
		logrus.Errorf("cluster may be unhealthy: failed to list members")
		return "", err
	}

	status := make(map[string][]string)
	for _, m := range ms {
		if len(m.ClientURLs) == 0 {
			logrus.Errorf("member %s is unreachable: no available published client urls\n", m.ID)
			continue
		}

		checked := false
		for _, url := range m.ClientURLs {
			resp, err := hc.Get(url + "/health")
			if err != nil {
				logrus.Errorf("failed to check the health of member %s on %s: %v\n", m.ID, url, err)
				continue
			}

			result := struct{ Health string }{}
			nresult := struct{ Health bool }{}
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Errorf("failed to check the health of member %s on %s: %v\n", m.ID, url, err)
				continue
			}
			resp.Body.Close()

			err = json.Unmarshal(bytes, &result)
			if err != nil {
				err = json.Unmarshal(bytes, &nresult)
			}
			if err != nil {
				logrus.Errorf("failed to check the health of member %s on %s: %v\n", m.ID, url, err)
				continue
			}

			checked = true
			if result.Health == "true" || nresult.Health == true {
				status[m.ID] = []string{"healthy", url}
				logrus.Infof("member %s is healthy: got healthy result from %s\n", m.ID, url)
			} else {
				status[m.ID] = []string{"unhealthy", url}
				logrus.Errorf("member %s is unhealthy: got unhealthy result from %s\n", m.ID, url)
			}

			break
		}
		if !checked {
			status[m.ID] = m.ClientURLs
			logrus.Errorf("member %s is unreachable: %v are all unreachable\n", m.ID, m.ClientURLs)
		}
	}

	logrus.Infof("item size = %d\n", len(status))
	ret, err := json.Marshal(status)
	if err != nil {
		logrus.Errorf("encode json error")
		return "", err
	}

	return string(ret), nil
}
