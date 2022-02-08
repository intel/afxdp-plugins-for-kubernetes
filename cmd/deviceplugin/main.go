/*
 * Copyright(c) 2022 Intel Corporation.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/intel/afxdp_k8s_plugins/internal/deviceplugin"
	"github.com/intel/afxdp_k8s_plugins/internal/host"
	"github.com/intel/afxdp_k8s_plugins/internal/logformats"
	"github.com/intel/afxdp_k8s_plugins/internal/networking"
	logging "github.com/sirupsen/logrus"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	defaultConfigFile = "./config.json"
	devicePrefix      = "cndp"
	minLinuxVersion   = "4.18.0" // Minimum Linux version for AF_XDP support
)

type devicePlugin struct {
	pools map[string]deviceplugin.PoolManager
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", defaultConfigFile, "Location of the device plugin configuration file")
	flag.Parse()

	logging.SetReportCaller(true)
	logging.SetFormatter(logformats.Default)

	logging.Infof("Starting CNDP Device Plugin")

	host := host.NewHandler()

	// kernel
	linuxVersion, err := host.KernelVersion()
	if err != nil {
		logging.Errorf("Error checking kernel version: %v", err)
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}
	logging.Infof("Kernel version: %v", linuxVersion)

	linuxInt, err := intVersion(linuxVersion)
	if err != nil {
		logging.Errorf("Error converting kernel version: %v", err)
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}

	minLinuxInt, err := intVersion(minLinuxVersion)
	if err != nil {
		logging.Errorf("Error converting kernel version: %v", err)
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}

	if linuxInt < minLinuxInt {
		logging.Errorf("Kernel version %v is below minimum requirement", linuxVersion)
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}

	// libbpf
	bpfInstalled, err := host.HasLibbpf()
	if err != nil {
		logging.Errorf("Error checking bpfInstalled: %v", err)
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}
	if bpfInstalled {
		logging.Infof("Libbpf present on host")
	} else {
		logging.Errorf("Libbpf not found on host")
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}

	// ethtool
	ethInstalled, err := host.HasEthtool()
	if err != nil {
		logging.Errorf("Error checking ethInstalled: %v", err)
	}
	if ethInstalled {
		logging.Infof("Ethtool present on host")
	} else {
		logging.Errorf("Ethtool not found on host")
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}

	// get config
	cfg, err := deviceplugin.GetConfig(configFile, networking.NewHandler())
	if err != nil {
		logging.Errorf("Error getting device plugin config: %v", err)
		logging.Errorf("Device plugin will exit")
		os.Exit(1)
	}

	// unprivileged_bpf_disabled
	unprivBpfAllowed, err := host.AllowsUnprivilegedBpf()
	if err != nil {
		logging.Errorf("Error checking if host allows Unprivileged BPF operations: %v", err)
	}
	if unprivBpfAllowed {
		logging.Infof("Unprivileged BPF is allowed")
	} else {
		logging.Warningf("Unprivileged BPF is disabled")
		if cfg.RequireUnprivilegedBpf {
			logging.Errorf("Unprivileged bpf is required")
			logging.Errorf("Device plugin will exit")
			os.Exit(1)
		}
	}

	dp := devicePlugin{
		pools: make(map[string]deviceplugin.PoolManager),
	}

	for _, poolConfig := range cfg.Pools {
		pm := deviceplugin.PoolManager{
			Name:          poolConfig.Name,
			Mode:          cfg.Mode,
			Devices:       make(map[string]*pluginapi.Device),
			DpAPISocket:   pluginapi.DevicePluginPath + devicePrefix + "-" + poolConfig.Name + ".sock",
			DpAPIEndpoint: devicePrefix + "-" + poolConfig.Name + ".sock",
			UpdateSignal:  make(chan bool),
			Timeout:       cfg.Timeout,
			DevicePrefix:  devicePrefix,
		}

		if err := pm.Init(poolConfig); err != nil {
			logging.Warningf("Error initializing pool: %v", pm.Name)
			logging.Errorf("%v", err)
		}

		dp.pools[poolConfig.Name] = pm
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sigs
	logging.Infof("Received signal \"%v\"", s)
	for _, pm := range dp.pools {
		logging.Infof("Terminating %v", pm.Name)
		if err := pm.Terminate(); err != nil {
			logging.Errorf("Termination error: %v", err)
		}
	}
}

func intVersion(version string) (int64, error) { // example "5.4.0-89-generic"
	stripped := strings.Split(version, "-")[0] // "5.4.0"
	split := strings.Split(stripped, ".")      // [5 4 0]

	padded := ""
	for _, val := range split { // 000500040000
		padded += fmt.Sprintf("%04s", val)
	}

	value, err := strconv.ParseInt(padded, 10, 64) // 500040000
	if err != nil {
		return -1, err
	}

	return value, nil
}
