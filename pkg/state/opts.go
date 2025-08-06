// SPDX-License-Identifier: Apache-2.0

package state

type StateOpt func(s *State)

// WithPgrollVersion sets the version of `pgroll` that is constructing the State
// instance
func WithPgrollVersion(version string) StateOpt {
	return func(s *State) {
		s.pgrollVersion = version
	}
}

// WithDryRun enables dry run mode for the State instance
func WithDryRun(enabled bool) StateOpt {
	return func(s *State) {
		s.dryRun = enabled
	}
}
