package policy

import "fmt"

type PolicyLoaderError struct {
	loaderError error
}

func (e *PolicyLoaderError) Error() string {
	return fmt.Sprintf("load: %v", e.loaderError)
}

type NoPoliciesError struct {
	policyPaths []string
}

func (e *NoPoliciesError) Error() string {
	return fmt.Sprintf("no policy .rego files found in %v", e.policyPaths)
}
