package cephclient

import (
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

type CephClient struct {
	monHost string
	monPort int
	conn    *rados.Conn
	ioctxs  map[string]*rados.IOContext
	images  map[string]map[string]*rbd.Image
}

/*
func NewCephClient(Host string, port int) (*CephClient, error) {
	var cephclient = &CephClient{monHost: Host, monPort: port}
	cephclient.ioctxs = make(map[string]*rados.IOContext)
	cephclient.images = make(map[string]map[string]*rbd.Image)
	err := cephclient.connect()
	if err != nil {
		return nil, err
	}
	return cephclient, nil

}
*/

func NewCephClient(args ...interface{}) (*CephClient, error) {
	if (len(args) != 0) && (len(args) != 2) {
		return nil, fmt.Errorf("input parameter number is error")
	}
	var host string
	var port int = -1
	for _, arg := range args {
		switch a := arg.(type) {
		case int:
			port = int(a)
		case int64:
			port = int(a)
		case int32:
			port = int(a)
		case string:
			host = a
		default:
			return nil, fmt.Errorf("input parameter type is error")
		}
	}
	var cephclient *CephClient = nil
	if (host != "") && (port != -1) {
		cephclient = &CephClient{monHost: host, monPort: port}
	} else {
		cephclient = &CephClient{}
	}
	cephclient.ioctxs = make(map[string]*rados.IOContext)
	cephclient.images = make(map[string]map[string]*rbd.Image)
	err := cephclient.connect()
	if err != nil {
		return nil, err
	}
	return cephclient, nil

}

func (cc *CephClient) connect() error {
	conn, err := rados.NewConn()
	if err != nil {
		return err
	}
	conn.ReadDefaultConfigFile()

	if cc.monHost != "" {
		monAddr := fmt.Sprintf("%s:%d", cc.monHost, cc.monPort)
		fmt.Println(monAddr)
		conn.SetConfigOption("mon_host", monAddr)
	}
	conn.SetConfigOption("client_mount_timeout", "5")
	err = conn.Connect()
	if err != nil {
		conn.Shutdown()
		return err
	}
	cc.conn = conn
	return nil
}

func (cc *CephClient) Destroy() {
	for pool, _ := range cc.ioctxs {
		cc.destroyPool(pool)
	}
	cc.conn.Shutdown()
	cc.conn = nil
}

func (cc *CephClient) ListPools() ([]string, error) {
	if cc.conn == nil {
		return nil, fmt.Errorf("can not connect to monitor node")
	}
	pools, err := cc.conn.ListPools()
	if err != nil {
		return nil, err
	} else {
		return pools, nil
	}
}

func (cc *CephClient) CreatePool(name string) error {
	err := cc.conn.MakePool(name)
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) DeletePool(name string) error {
	cc.destroyPool(name)
	err := cc.conn.DeletePool(name)
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) openPool(name string) (*rados.IOContext, error) {
	ioctx, ok := cc.ioctxs[name]
	if ok {
		return ioctx, nil
	}

	ioctx, err := cc.conn.OpenIOContext(name)
	if err != nil {
		return nil, err
	}
	cc.ioctxs[name] = ioctx
	return ioctx, nil
}

func (cc *CephClient) destroyPool(name string) {
	cc.closeImages(name)
	ioctx, ok := cc.ioctxs[name]
	if !ok {
		return
	} else {
		ioctx.Destroy()
		delete(cc.ioctxs, name)
		return
	}
}

func (cc *CephClient) ListImageNames(poolName string) ([]string, error) {
	ioctx, err := cc.openPool(poolName)
	if err != nil {
		return nil, err
	}
	imageNames, err := rbd.GetImageNames(ioctx)
	if err != nil {
		return nil, err
	} else {
		return imageNames, nil
	}
}

// size单位为M
func (cc *CephClient) CreateImage(poolName string, imageName string,
	size uint64, order int) error {
	ioctx, err := cc.openPool(poolName)
	if err != nil {
		return err
	}
	size = size << 20
	var features uint64 = 1
	_, err = rbd.Create(ioctx, imageName, size, order, features)
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) DeleteImage(poolName string, imageName string) error {
	image, err := cc.openImage(poolName, imageName)
	if err != nil {
		return err
	}
	cc.closeImage(poolName, imageName)
	err = image.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) openImage(poolName string, imageName string) (*rbd.Image, error) {
	images, ok := cc.images[poolName]
	if ok {
		image, ok := images[imageName]
		if ok {
			return image, nil
		}
	}
	ioctx, err := cc.openPool(poolName)
	if err != nil {
		return nil, err
	}
	img := rbd.GetImage(ioctx, imageName)
	err = img.Open()
	if err != nil {
		return nil, err
	}

	images, ok = cc.images[poolName]
	if ok {
		images[imageName] = img
	} else {
		images = map[string]*rbd.Image{imageName: img}
		cc.images[poolName] = images
	}
	return img, nil
}

func (cc *CephClient) closeImage(poolName string, imageName string) {
	images, ok := cc.images[poolName]
	if ok {
		image, ok := images[imageName]
		if ok {
			image.Close()
			delete(images, imageName)
			if len(images) == 0 {
				delete(cc.images, poolName)
			}
			return
		}
	}
}

func (cc *CephClient) closeImages(poolName string) {
	images, ok := cc.images[poolName]
	if ok {
		for _, image := range images {
			image.Close()
		}
		delete(cc.images, poolName)
		return
	}
}

func (cc *CephClient) ImageStat(poolName string, imageName string) (map[string]uint64, error) {
	image, err := cc.openImage(poolName, imageName)
	if err != nil {
		return nil, err
	}
	imginfo, err := image.Stat()
	if err != nil {
		return nil, err
	} else {
		imgInfoMap := make(map[string]uint64, 2)
		imgInfoMap["size"] = imginfo.Size >> 20
		imgInfoMap["obj_size"] = imginfo.Obj_size >> 20
		return imgInfoMap, nil
	}
}

func (cc *CephClient) CreateSnapshot(poolName string, imageName string, snapName string) error {
	image, err := cc.openImage(poolName, imageName)
	if err != nil {
		return err
	}
	_, err = image.CreateSnapshot(snapName)
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) getSnapshot(poolName string, imageName string, snapName string) (*rbd.Snapshot, error) {
	image, err := cc.openImage(poolName, imageName)
	if err != nil {
		return nil, err
	}
	snapshot := image.GetSnapshot(snapName)
	return snapshot, nil
}

func (cc *CephClient) RemoveSnapshot(poolName string, imageName string, snapName string) error {
	snapshot, err := cc.getSnapshot(poolName, imageName, snapName)
	if err != nil {
		return err
	}
	err = snapshot.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) Rollback(poolName string, imageName string, snapName string) error {
	snapshot, err := cc.getSnapshot(poolName, imageName, snapName)
	if err != nil {
		return err
	}
	err = snapshot.Rollback()
	if err != nil {
		return err
	}
	return nil
}

func (cc *CephClient) GetSnapshotNames(poolName string, imageName string) ([]string, error) {
	image, err := cc.openImage(poolName, imageName)
	if err != nil {
		return nil, err
	}
	snapInfo, err := image.GetSnapshotNames()
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v", snapInfo)
	snapNames := make([]string, len(snapInfo))
	for index, snap := range snapInfo {
		snapNames[index] = snap.Name
	}
	return snapNames, nil
}

func (cc *CephClient) Backup(poolName string, imageName string, snapName string, destImageName string) error {

	ioctx, err := cc.openPool(poolName)
	if err != nil {
		return err
	}
	image, err := cc.openImage(poolName, imageName)
	if err != nil {
		return err
	}
	snapshot, err := cc.getSnapshot(poolName, imageName, snapName)
	if err != nil {
		return err
	}
	err = snapshot.Protect()
	if err != nil {
		return err
	}
	defer snapshot.Unprotect()

	var features uint64 = 1
	var order int = 22

	_, err = image.Clone(snapName, ioctx, destImageName, features, order)
	if err != nil {
		return err
	}
	destImage, err := cc.openImage(poolName, destImageName)
	if err != nil {
		return err
	}
	err = destImage.Flatten()
	if err != nil {
		return err
	}
	return nil
}
