// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

//go:build !unix

package feed

import "time"

// CPUTime is not measured on this platform; the driver prints n/a.
func CPUTime() (time.Duration, bool) { return 0, false }
