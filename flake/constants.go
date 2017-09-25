package flake

// OvertoneEpochMs represents the number of milliseconds that elapsed between the
// UnixEpoch (1/1/1970) and the Overtone Epoch, Sunday, January 1st, 2017
// (2017-01-01 00:00:00 +0000 UTC)
//
// This is the default epoch used by overt-flake
const OvertoneEpochMs = int64(1483228800000)

// UnixEpochMs is the number of milliseconds elapsed since the Unix Epoch
// (which is none)
const UnixEpochMs = int64(0)

// SnowflakeEpochMs is the number of milliseconds elapsed between the Unix Epoch
// (1/1/19070) and the Twitter Snowflake Epoch (2010-11-04 01:42:54 +0000 UTC ??)
const SnowflakeEpochMs = int64(1288834974657)

// OvertFlakeIDLength is the length, in bytes, of an Overt-flake ID
const OvertFlakeIDLength = 16

// MACAddressLength is the length, in bytes, of a Network MAC address
const MACAddressLength = 6
