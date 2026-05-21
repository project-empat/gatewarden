package core

const Version = "0.1.0"
const ProductName = "Gatewarden"

type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type CheckFunc func() (*CheckResult, error)
