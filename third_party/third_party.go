package thirdparty

import (
	_ "embed"
)

// ThirdPartyLicenses embeds the VENDOR-LICENSE file which includes all 3P licenses for Kf.
//go:embed VENDOR-LICENSE
var ThirdPartyLicenses string

