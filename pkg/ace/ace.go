package ace

//#cgo CFLAGS: -I../../include
//#cgo pkg-config: ace_bt
//#include "ace.h"
import "C"
import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"unsafe"

	"github.com/clintharrison/kbt-cgo/pkg/withlock"
)

// Unfortunately these have to be globals because the C callbacks need to access them.
var (
	mu                 sync.Mutex
	sessionHandle      C.aceBT_sessionHandle
	scanInstanceHandle C.aceBT_scanInstanceHandle
	scanningAdapter    *AceAdapter
	scanResultFunc     func(adapter *AceAdapter, device ScanResult)
)

type AceAdapter struct{}

func (a *AceAdapter) Scan(f func(adapter *AceAdapter, device ScanResult)) error {
	err := withlock.DoErr(&mu, func() error {
		if scanInstanceHandle != nil {
			return errors.New("scan already in progress")
		}
		scanResultFunc = f
		return nil
	})

	if err != nil {
		slog.Error("Failed to start scan", "error", err)
		return err
	}
	client_id := (C.aceBT_BeaconClientId)(C.ACE_BEACON_CLIENT_TYPE_MONEYPENNY)
	aceStatus := C.aceBT_startBeaconScanWithDefaultParams(sessionHandle, client_id, &scanInstanceHandle)
	if err := errForStatus(aceStatus); err != nil {
		slog.Error("Failed to start beacon scan", "status", aceStatus, "error", err)
		return err
	}
	return nil
}

func (a *AceAdapter) StopScan() error {
	if scanInstanceHandle == nil {
		return errors.New("no scan in progress")
	}
	aceStatus := C.aceBT_stopBeaconScan(scanInstanceHandle)
	if err := errForStatus(aceStatus); err != nil {
		slog.Error("Failed to stop beacon scan", "status", aceStatus, "error", err)
		return err
	}
	slog.Info("Stopped beacon scan")
	scanInstanceHandle = nil
	return nil
}

type Address struct {
	addr [6]uint8
}

func NewAddressFromAce(addr C.aceBT_bdAddr_t) Address {
	bs := addr.address
	return Address{addr: [6]uint8{uint8(bs[0]), uint8(bs[1]), uint8(bs[2]), uint8(bs[3]), uint8(bs[4]), uint8(bs[5])}}
}

func (a Address) ToAce() *C.aceBT_bdAddr_t {
	addr := &C.aceBT_bdAddr_t{}
	for i := 0; i < 6; i++ {
		addr.address[i] = C.uint8_t(a.addr[i])
	}
	return addr
}

func (a Address) ToString() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		a.addr[0], a.addr[1], a.addr[2],
		a.addr[3], a.addr[4], a.addr[5])
}

type ScanResult struct {
	// The raw record, used by aceBT_scanRecordExtract* functions
	record *C.aceBT_BeaconScanRecord_t
	// Device address
	addr Address
	// RSSI of the remote advertisement
	rssi C.int
}

func (sr *ScanResult) Name() string {
	var name C.aceBT_bdName_t
	nameLen := C.aceBT_scanRecordExtractName(sr.record, &name)
	if nameLen > 0 {
		nameBytes := C.GoBytes(unsafe.Pointer(&name.name[0]), C.int(nameLen))
		return string(nameBytes)
	}
	return "<unknown>"
}

func (sr *ScanResult) Address() Address {
	return sr.addr
}

func (sr *ScanResult) RSSI() int {
	return int(sr.rssi)
}

func (sr *ScanResult) TxPower() int {
	var txPower C.int
	len := C.aceBT_scanRecordExtractTxPower(sr.record, &txPower)
	if len == 1 {
		return int(txPower)
	}
	return 0
}

func errForStatus(status C.ace_status_t) error {
	switch status {
	case C.ACE_STATUS_OK:
		return nil
	case C.ACEBT_STATUS_NOMEM:
		return errors.New("ACE out of memory")
	case C.ACEBT_STATUS_BUSY:
		return errors.New("ACE is busy connecting another device")
	case C.ACEBT_STATUS_PARM_INVALID:
		return errors.New("ACE request contains invalid parameters")
	case C.ACEBT_STATUS_NOT_READY:
		return errors.New("ACE server not ready")
	case C.ACEBT_STATUS_FAIL:
		return errors.New("ACE failed")
	default:
		return errors.New(fmt.Sprintf("ACE unknown error: %d", status))
	}
}

func InitAdapterWithSession() (*AceAdapter, error) {
	adapter := &AceAdapter{}
	if err := adapter.OpenSession(); err != nil {
		return nil, err
	}

	state, err := adapter.RadioState()
	if err != nil {
		slog.Error("Failed to get radio state", "error", err)
		return nil, err
	}
	if state != RadioEnabled {
		slog.Info("Radio is not enabled", "state", state)
		err = adapter.EnableRadio()
		if err != nil {
			slog.Error("Failed to enable radio", "error", err)
			return nil, err
		}
	}
	return adapter, nil
}

func (a *AceAdapter) OpenSession() error {
	session_type := (C.aceBT_sessionType_t)(C.ACEBT_SESSION_TYPE_BLE)
	status := C.aceBT_openSession(session_type, nil, &sessionHandle)
	if err := errForStatus(status); err != nil {
		slog.Error("Failed to open ACE session", "status", status, "error", err)
		return err
	}
	slog.Info("Opened ACE session", "sessionHandle", fmt.Sprintf("%p", sessionHandle))
	return nil
}

type AceRadioState int

const (
	RadioDisabled AceRadioState = iota
	RadioEnabled
	RadioEnabling
	RadioDisabling
)

func (a *AceAdapter) RadioState() (AceRadioState, error) {
	var radioState C.aceBT_state_t
	bleStatus := C.aceBT_getRadioState(&radioState)
	if err := errForStatus(bleStatus); err != nil {
		slog.Error("Failed to get radio state", "status", bleStatus, "error", err)
		return RadioDisabled, err
	}
	switch radioState {
	case C.ACEBT_STATE_DISABLED:
		return RadioDisabled, nil
	case C.ACEBT_STATE_ENABLED:
		return RadioEnabled, nil
	case C.ACEBT_STATE_ENABLING:
		return RadioEnabling, nil
	case C.ACEBT_STATE_DISABLING:
		return RadioDisabling, nil
	default:
		return RadioDisabled, fmt.Errorf("unknown radio state: %d", radioState)
	}
}

const readyWaitDelay = 500 * time.Millisecond

func (a *AceAdapter) EnableRadio() error {
	maxRetries := 10
	if sessionHandle == nil {
		return errors.New("session handle is nil, cannot enable radio")
	}
	slog.Info("Enabling radio", "sessionHandle", fmt.Sprintf("%p", sessionHandle))

	if err := errForStatus(C.aceBT_enableRadio(sessionHandle)); err != nil {
		slog.Error("failed to enable radio", "error", err)
	}

	for i := 0; i < maxRetries; i++ {
		radioState, err := a.RadioState()
		if err != nil {
			slog.Error("failed to get radio state", "error", err)
			return err
		}
		if radioState == RadioEnabled {
			slog.Debug("radio is enabled, quitting retry loop")
			return nil
		}
		time.Sleep(readyWaitDelay)
	}
	return fmt.Errorf("radio did not enable after %d retries", maxRetries)
}

func (a *AceAdapter) RegisterBeacon() error {
	bleStatus := C.aceBT_RegisterBeaconClient(
		sessionHandle,
		&C.beacon_callbacks,
	)
	if err := errForStatus(bleStatus); err != nil {
		slog.Error("Failed to register beacon client", "status", bleStatus, "error", err)
		return err
	}
	return nil
}

//export advChangeCallback
func advChangeCallback(adv_instance C.aceBT_advInstanceHandle, state C.aceBT_beaconAdvState_t, power_mode C.aceBT_beaconPowerMode_t, beacon_mode C.aceBT_beaconAdvMode_t) {
	slog.Info("Beacon advertisement state changed", "adv_instance", adv_instance, "state", state, "power_mode", power_mode, "beacon_mode", beacon_mode)
}

//export scanResultCallback
func scanResultCallback(scan_instance C.aceBT_scanInstanceHandle, record *C.aceBT_BeaconScanRecord_t) {
	scanResultFunc(scanningAdapter, ScanResult{
		record: record,
		addr:   NewAddressFromAce(record.addr),
		rssi:   record.rssi,
	})
}

//export scanChangeCallback
func scanChangeCallback(scan_instance C.aceBT_scanInstanceHandle, state C.aceBT_beaconScanState_t, interval uint32, window uint32) {
	stateStr := "unknown"
	switch state {
	case C.ACEBT_BEACON_SCAN_FAILED:
		stateStr = "failed"
	case C.ACEBT_BEACON_SCAN_QUEUED:
		stateStr = "queued"
	case C.ACEBT_BEACON_SCAN_STARTED:
		stateStr = "started"
	case C.ACEBT_BEACON_SCAN_PAUSED:
		stateStr = "paused"
	case C.ACEBT_BEACON_SCAN_STOPPED:
		stateStr = "stopped"
	}
	slog.Info("Beacon scan state changed", "state", stateStr, "interval", interval, "window", window)
}

//export onBeaconClientRegistered
func onBeaconClientRegistered(status C.ace_status_t) {
	if err := errForStatus(status); err != nil {
		slog.Error("Beacon client registration failed", "status", status, "error", err)
		return
	}
	slog.Info("Beacon client registered successfully", "status", status)
}
