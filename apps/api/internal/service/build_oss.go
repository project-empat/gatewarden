//go:build !enterprise

package service

// buildMode reports the active build flavor. Enterprise builds are produced
// with GOWORK=go.work.enterprise and the enterprise Go build tag.
const buildMode = "oss"
