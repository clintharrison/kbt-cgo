#include "ace.h"

aceBT_beaconCallbacks_t beacon_callbacks = {
    .size = sizeof(aceBT_beaconCallbacks_t),
    // Advertisement state changed
    .advStateChanged = advChangeCallback,
    // Scan state changed
    .scanStateChanged = scanChangeCallback,
    // Scan results callback
    .scanResults = scanResultCallback,
    // Beacon client registration callback
    .onclientRegistered = onBeaconClientRegistered,
};