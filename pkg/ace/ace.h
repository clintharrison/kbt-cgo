#pragma once

#include "ace/ace_status.h"
#include "ace/bluetooth_session_api.h"
#include "ace/bluetooth_common_api.h"
#include "ace/bluetooth_beacon_api.h"

/**
 * @brief callback to notifiy a change in advertisment instance\n
 * Invoked on @ref aceBT_startBeacon, @ref aceBT_startBeaconWithScanResponse,
 * and @ref aceBT_stopBeacon
 *
 * @param[in] adv_instance Advertisement instance
 * @param[in] state Current advertisement state
 * @param[in] power_mode Current power mode used for this advertisement
 * @param[in] beacon_mode Beacon mode in which this adv instance is being broadcasted
 *     typedef void (*beacon_advChangeCallback)(aceBT_advInstanceHandle adv_instance,
 *                                              aceBT_beaconAdvState_t state,
 *                                              aceBT_beaconPowerMode_t power_mode,
 *                                              aceBT_beaconAdvMode_t beacon_mode);
 */
extern void advChangeCallback(aceBT_advInstanceHandle adv_instance, aceBT_beaconAdvState_t state, aceBT_beaconPowerMode_t power_mode, aceBT_beaconAdvMode_t beacon_mode);

/**
 * @brief callback to notifiy a change in advertisment instance\n
 * Invoked on @ref aceBT_startBeaconScan, @ref
 * aceBT_startBeaconScanWithDefaultParams, @ref aceBT_stopBeaconScan
 *
 * @param[in] scan_instance Scan instance
 * @param[in] state Current advertisement state
 * @param[in] interval Interval in in untis of 1.25 ms at which this scan is
 * performed currently
 * @param[in] window length of scan procedure / scan interval in untis of 1.25
 * ms
 *     typedef void (*beacon_scanChangeCallback)(
 *         aceBT_scanInstanceHandle scan_instance, aceBT_beaconScanState_t state,
 *         uint32_t interval, uint32_t window);
 */
extern void scanChangeCallback(aceBT_scanInstanceHandle scan_instance, aceBT_beaconScanState_t state, uint32_t interval, uint32_t window);

/**
 * @brief callback to notifiy a change in advertisment instance\n
 * Invoked in response of @ref aceBT_startBeaconScan and @ref
 * aceBT_startBeaconScanWithDefaultParams
 *
 * @param[in] scan_instance Scan instance
 * @param[in] state Current advertisement state
 * @param[in] scanResult Scan result(s)
 */
extern void scanResultCallback(aceBT_scanInstanceHandle scan_instance, aceBT_BeaconScanRecord_t* record);

/**
 * @brief callback to notifiy that beacon client registration status\n
 * Invoked on @ref aceBT_RegisterBeaconClient
 *
 * @param[in] status status of the beacon client registration
 *     typedef void (*beacon_onBeaconClientRegistered)(aceBT_status_t status);
 */
extern void onBeaconClientRegistered(aceBT_status_t status);

// these are passed to aceBT_RegisterBeaconClient
extern aceBT_beaconCallbacks_t beacon_callbacks;