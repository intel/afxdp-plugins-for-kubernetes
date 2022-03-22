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
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/deviceplugin"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/logformats"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	logging "github.com/sirupsen/logrus"
	"io"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	exitNormal        = 0
	exitConfigError   = 1
	exitLogError      = 2
	exitHostError     = 3
	exitPoolError     = 4
	defaultConfigFile = "./config.json"
	devicePrefix      = "cndp"
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

	// config
	cfg, err := deviceplugin.GetConfig(configFile, networking.NewHandler())
	if err != nil {
		logging.Errorf("Error getting device plugin config: %v", err)
		exit(exitConfigError)
	}

	// logging
	if err := configureLogging(cfg); err != nil {
		logging.Errorf("Error configuring logging: %v", err)
		exit(exitLogError)
	}

	logging.Infof("Starting AF_XDP Device Plugin")
	logging.Infof("Device Plugin mode: %s", cfg.Mode)

	// requirements
	logging.Infof("Checking if host meets requriements")
	hostMeetsRequirements, err := checkHost(host.NewHandler(), cfg)
	if err != nil {
		logging.Errorf("Error checking host post config: %v", err)
		exit(exitHostError)
	}
	if !hostMeetsRequirements {
		logging.Infof("Host does not meet requriements")
		exit(exitNormal)
	}
	logging.Infof("Host meets requriements")

	// pools
	logging.Infof("Building device pools")
	if err := cfg.BuildPools(); err != nil {
		logging.Warningf("Error building device pools: %v", err)
		exit(exitPoolError)
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
			Timeout:       cfg.UdsTimeout,
			DevicePrefix:  devicePrefix,
			CndpFuzzTest:  cfg.CndpFuzzing,
		}

		if err := pm.Init(poolConfig); err != nil {
			logging.Errorf("Error initializing pool %v: %v", pm.Name, err)
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

func configureLogging(cfg deviceplugin.Config) error {
	err := os.MkdirAll(cfg.LogDir, cfg.LogDirPermission)
	if err != nil {
		logging.Errorf("Error setting log directory: %v", err)
	}

	if cfg.LogFile != "" {
		logging.Infof("Setting log file: %s", cfg.LogFile)
		fp, err := os.OpenFile(cfg.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, cfg.LogFilePermission)
		if err != nil {
			logging.Errorf("Error setting log file: %v", err)
			return err
		}
		logging.SetOutput(io.MultiWriter(fp, os.Stdout))
	}

	if cfg.LogLevel != "" {
		logging.Infof("Setting log level: %s", cfg.LogLevel)
		level, err := logging.ParseLevel(cfg.LogLevel)
		if err != nil {
			logging.Errorf("Error setting log level: %v", err)
			return err
		}
		logging.SetLevel(level)

		if cfg.LogLevel == "debug" {
			logging.Infof("Switching to debug log format")
			logging.SetFormatter(logformats.Debug)
		}
	}

	return nil
}

func checkHost(host host.Handler, cfg deviceplugin.Config) (bool, error) {
	// kernel
	logging.Debugf("Checking kernel version")
	linuxVersion, err := host.KernelVersion()
	if err != nil {
		err := fmt.Errorf("Error checking kernel version: %v", err)
		return false, err

	}

	linuxInt, err := intVersion(linuxVersion)
	if err != nil {
		err := fmt.Errorf("Error converting actual kernel version to int: %v", err)
		return false, err

	}

	minLinuxInt, err := intVersion(cfg.MinLinuxVersion)
	if err != nil {
		err := fmt.Errorf("Error converting minimum kernel version to int: %v", err)
		return false, err

	}

	if linuxInt < minLinuxInt {
		logging.Warningf("Kernel version %v is below minimum requirement %v", linuxVersion, cfg.MinLinuxVersion)
		return false, nil
	}
	logging.Debugf("Kernel version: %v meets minimum requirements", linuxVersion)

	// libbpf
	logging.Debugf("Checking host for Libbpf")
	bpfInstalled, libs, err := host.HasLibbpf()
	if err != nil {
		err := fmt.Errorf("Libbpf not found on host")
		return false, err
	}
	if bpfInstalled {
		logging.Debugf("Libbpf found on host:")
		for _, lib := range libs {
			logging.Debugf("\t" + lib)
		}
	} else {
		logging.Warningf("Libbpf not found on host")
		return false, nil
	}

	// unprivileged bpf
	logging.Debugf("Checking if host allows unprivileged BPF operations")
	unprivBpfAllowed, err := host.AllowsUnprivilegedBpf()
	if err != nil {
		logging.Errorf("Error checking if host allows unprivileged BPF operations: %v", err)
		return false, err
	}
	if unprivBpfAllowed {
		logging.Debugf("Unprivileged BPF is allowed")
	} else {
		logging.Warningf("Unprivileged BPF is disabled")
		if cfg.RequireUnprivilegedBpf {
			logging.Warningf("Unprivileged BPF is required")
			return false, nil
		}
	}

	// ethtool
	logging.Debugf("Checking host for Ethtool")
	ethInstalled, version, err := host.HasEthtool()
	if err != nil {
		logging.Errorf("Error checking if Ethtool is present on host: %v", err)
		return false, err
	}
	if ethInstalled {
		logging.Debugf("Ethtool found on host:")
		logging.Debugf("\t" + version)
	} else {
		logging.Warningf("Ethool not found on host")
		return false, nil
	}

	return true, nil
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

func exit(code int) {
	if code == 0 {
		logging.Infof("Device plugin will exit")
	} else {
		logging.Errorf("Device plugin will exit")
	}
	os.Exit(code)
}
