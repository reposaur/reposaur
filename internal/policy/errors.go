package policy

import "fmt"

type ErrPolicyLoad struct {
	loaderError error
}

func (e *ErrPolicyLoad) Error() string {
	return fmt.Sprintf("load: %v", e.loaderError)
}

type ErrNoPolicies struct {
	policyPaths []string
}

func (e *ErrNoPolicies) Error() string {
	return fmt.Sprintf("no policy .rego files found in %v", e.policyPaths)
}
