/*
 * Copyright(c) Red Hat
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	 http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bpf

import (
	"fmt"
	"os"
	"syscall"

	"github.com/google/uuid"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/moby/sys/mount"
	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
)

const (
	pinnedMapBaseDir     = "/var/run/afxdp_dp/"
	pinnedMapDirFileMode = os.FileMode(0755)
	bpffsDirFileMode     = os.FileMode(0755)
)

/*
MapManager is the interface defining the MAP MANAGER.
Implementations of this interface are the main type of this MapManager package. TODO UPDATE
*/
type MapManager interface {
	CreateBPFFS(dev, path string) (string, error)
	DeleteBPFFS(dev string) error
	AddMap(dev, path string)
	GetMaps() map[string]string
	GetBPFFS(dev string) string
}

type PoolBpfMapManager struct {
	Manager MapManager
	Path    string
}

/*
MapManagerFactory is the interface defining a factory that creates and returns MapManagers.
Each device plugin poolManager will have its own MapManagerFactory and each time a
container is created the factory will create a MapManager to serve the
associated pinned BPF Map. TODO UPDATE THIS....
*/
type MapManagerFactory interface {
	CreateMapManager(poolName, user string) (MapManager, string, error)
}

/*
server implements the Server interface. It is the main type for this package.
*/
type mapManager struct {
	name      string
	maps      map[string]string
	bpffsPath string
	uid       string
}

/*
mapManager implements the MapManager interface.
*/
type mapManagerFactory struct {
	MapManagerFactory
}

/*
NewMapMangerFactory returns an implementation of the MapManagerFactory interface.
*/
func NewMapMangerFactory() MapManagerFactory {
	return &mapManagerFactory{}
}

/*
CreateMapManager creates, initialises, and returns an implementation of the MapManager interface.
It also returns the filepath for bpf maps to be pinned.
*/
func (f *mapManagerFactory) CreateMapManager(poolName, user string) (MapManager, string, error) {

	logging.Debugf("	  CreateMapManager	  ")
	if poolName == "" || user == "" {
		return nil, "", errors.New("Error poolname or user not set")
	}
	p, err := createBPFFSBaseDirectory(poolName, user)
	if err != nil {
		return nil, "", errors.Wrapf(err, "Error creating BPFFS base directory %v", err.Error())
	}
	logging.Infof("Created BPFFS Base directory %s", p)

	manager := &mapManager{
		maps:      make(map[string]string),
		bpffsPath: p,
		uid:       user,
		name:      poolName,
	}

	return manager, p, nil
}

func giveBpffsBasePermissions(path, user string) error {
	if user != "0" {
		logging.Infof("Giving permissions to UID %s", user)
		err := host.GivePermissions(path, user, "rwx")
		if err != nil {
			return errors.Wrapf(err, "Error giving permissions to BPFFS path %s", err.Error())
		}
		logging.Infof("User %s has access to %s", user, path)
	}
	return nil
}

func createBPFFSBaseDirectory(p, user string) (string, error) {

	logging.Infof("Creating BPFFS Base directory %s", p)

	path := pinnedMapBaseDir + p
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//create base directory if it not exists, with correct file permissions
		if err = os.MkdirAll(path, pinnedMapDirFileMode); err != nil {
			return "", errors.Wrapf(err, "Error creating BPFFS base directory %s: %v", pinnedMapBaseDir, err.Error())
		}

		if err = giveBpffsBasePermissions(path, user); err != nil {
			return "", errors.Wrapf(err, "Error creating BPFFS base directory %s: %v", pinnedMapBaseDir, err.Error())
		}
	}

	logging.Infof("Created Base BPFFS directory %s", path)
	return path, nil
}

func (m mapManager) CreateBPFFS(device, path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.Wrapf(err, "Error creating BPFFS mount point base directory %s doesn't exist: %v", pinnedMapBaseDir, err.Error())
	}
	//TODO ADD DEVICE NAME
	bpffsPath, err := generateRandomBpffsName(m.bpffsPath)
	if err != nil {
		return "", errors.Wrapf(err, "Error generating BPFFS path: %s: %v", pinnedMapBaseDir, err.Error())
	}

	if err = os.MkdirAll(bpffsPath, bpffsDirFileMode); err != nil {
		return "", errors.Wrapf(err, "Error creating BPFFS base directory %s: %v", pinnedMapBaseDir, err.Error())
	}

	if err = giveBpffsBasePermissions(bpffsPath, m.uid); err != nil {
		return "", errors.Wrapf(err, "Error creating BPFFS base directory %s: %v", pinnedMapBaseDir, err.Error())
	}
	logging.Infof("created a directory %s", bpffsPath)

	if err = syscall.Mount(bpffsPath, bpffsPath, "bpf", 0, ""); err != nil {
		return "", errors.Wrapf(err, "failed to mount %s: %v", bpffsPath, err.Error())
	}
	logging.Infof("Created BPFFS mount point at %s", bpffsPath)

	if err = mount.MakeShared(bpffsPath); err != nil {
		return "", errors.Wrapf(err, "failed to make the BPFFS  %s Shared: %v", bpffsPath, err.Error())
	}

	return bpffsPath, nil
}

/*
generateRandomBpffsName will take the file directory path, and apply a unique name per each
bpffs created.
*/
func generateRandomBpffsName(directory string) (string, error) {

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return "", errors.Wrapf(err, "Error couldn't find directory %s: %v", directory, err.Error())
	}

	//get directory info
	fileInfo, err := os.Stat(directory)
	if err != nil {
		logging.Errorf("Error getting directory info %s: %v", directory, err)
		return "", err
	}

	//verify it is a directory
	if !fileInfo.IsDir() {
		err = fmt.Errorf("%s is not a directory", directory)
		logging.Errorf(err.Error())
		return "", err
	}

	//verify the permissions are correct, in case of pre existing dir
	if fileInfo.Mode().Perm() != bpffsDirFileMode {
		err = fmt.Errorf("incorrect permissions on directory %s", directory)
		logging.Errorf(err.Error())
		return "", err
	}

	var bpffspath string
	var count int = 0
	for {
		if count >= 5 {
			err = fmt.Errorf("error generating a unique UDS filepath")
			logging.Errorf(err.Error())
			return "", err
		}

		bpffsName, err := uuid.NewRandom()
		if err != nil {
			logging.Errorf("Error generating random UDS filename: %v", err)
		}

		bpffspath = directory + bpffsName.String()
		if _, err := os.Stat(bpffspath); os.IsNotExist(err) {
			break
		}

		logging.Debugf("%s already exists. Regenerating.", bpffspath)
		count++
	}

	return bpffspath, nil
}

/*
AddMap appends a netdev and its associated pinned xsk_map to the MapManager map of Maps.
*/
func (m *mapManager) AddMap(dev, path string) {
	m.maps[dev] = path
}

/*
GetMaps
*/
func (m *mapManager) GetMaps() map[string]string {
	return m.maps
}

/*
GetBPFFS
*/
func (m *mapManager) GetBPFFS(dev string) string {

	if p, ok := m.maps[dev]; ok {
		return p
	}

	return ""
}

func (m *mapManager) DeleteBPFFS(dev string) error {

	var err error

	bpffs := m.GetBPFFS(dev)
	if bpffs == "" {
		return errors.New("Could not find BPFFS")
	}

	if _, err := os.Stat(bpffs); os.IsNotExist(err) {
		return errors.Wrapf(err, "Error finding BPFFS directory %s doesn't exist: %v", bpffs, err.Error())
	}

	if err = syscall.Unmount(bpffs, 0); err != nil {
		return errors.Wrapf(err, "failed to umount %s: %v", bpffs, err.Error())
	}

	if err := os.Remove(bpffs); err != nil {
		return errors.Wrapf(err, "Error Remove BPFFS directory %s: %v", bpffs, err.Error())
	}

	logging.Infof("Deleted BPFFS mount point at %s", bpffs)

	return nil
}
