package main

import (
	"log/slog"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/lmittmann/tint"

	"github.com/clintharrison/kbt-cgo/pkg/ace"
)

const (
	BLUETOOTH_UID = 1003
	BLUETOOTH_GID = 1003
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// dropPrivileges sets the process's user and group ID to the Bluetooth user and group
// ACE will not allow the process to run as root.
func dropPrivileges() {
	if os.Geteuid() == 0 {
		err := syscall.Setgid(BLUETOOTH_GID)
		if err != nil {
			slog.Error("Failed to set GID", "error", err)
			os.Exit(1)
		}

		err = syscall.Setuid(BLUETOOTH_UID)
		if err != nil {
			slog.Error("Failed to set UID", "error", err)
			os.Exit(1)
		}
	}

	uid := syscall.Getuid()
	gid := syscall.Getgid()
	slog.Info("running as nonroot user", "uid", uid, "gid", gid)
}

// configureLogger sets up the default structured logger to use tint on stderr
func configureLogger() {
	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
}

func main() {
	configureLogger()
	dropPrivileges()

	slog.Info("Go Version", "version", runtime.Version(), "hostname", must(os.Hostname()))

	adapter, err := ace.InitAdapterWithSession()
	if err != nil {
		slog.Error("Failed to initialize ACE adapter", "error", err)
		os.Exit(1)
	}

	if err := adapter.RegisterBeacon(); err != nil {
		slog.Error("Failed to register beacon", "error", err)
		os.Exit(1)
	}

	devicesSeen := make(map[ace.Address]bool)
	deviceFoundChan := make(chan struct{})

	err = adapter.Scan(func(adapter *ace.AceAdapter, device ace.ScanResult) {
		if _, ok := devicesSeen[device.Address()]; ok {
			// quietly ignore devices we've already seen
			return
		}
		devicesSeen[device.Address()] = true

		if device.Name() == "<unknown>" {
			slog.Debug("found unnamed device", "address", device.Address().ToString(), "rssi", device.RSSI(), "tx_power", device.TxPower())
		} else if strings.HasPrefix(device.Name(), "Lightblue") {
			slog.Info("found Lightblue device, stopping scan", "address", device.Address().ToString(), "rssi", device.RSSI(), "tx_power", device.TxPower())
			// Stop the scan after finding the device
			if err := adapter.StopScan(); err != nil {
				slog.Error("failed to stop scan", "error", err)
			}
			close(deviceFoundChan)
		} else {
			slog.Info("found device", "name", device.Name(), "address", device.Address().ToString(), "rssi", device.RSSI(), "tx_power", device.TxPower())
		}
	})
	if err != nil {
		slog.Error("Failed to start scan", "error", err)
		os.Exit(1)
	}

	// Wait for a device to be found or timeout after 10 seconds
	select {
	case <-deviceFoundChan:
		// channel closed and device found
	case <-time.After(10 * time.Second):
		slog.Info("No device found within 10 seconds, stopping scan")
		if err := adapter.StopScan(); err != nil {
			slog.Error("Failed to stop scan", "error", err)
		}
	}
}
