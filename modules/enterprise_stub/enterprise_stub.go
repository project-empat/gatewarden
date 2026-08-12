//go:build !enterprise

package enterprise

import "errors"

var ErrEnterpriseOnly = errors.New("enterprise feature: not available in OSS build")

const BuildMode = "oss"

type EnterpriseFeature struct {
	Name string
}

func GetFeatures() []EnterpriseFeature {
	return nil
}
