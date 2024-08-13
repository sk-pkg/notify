// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import "strings"

func SpliceStr(p ...string) string {
	var b strings.Builder

	l := len(p)
	for i := 0; i < l; i++ {
		b.WriteString(p[i])
	}

	return b.String()
}
