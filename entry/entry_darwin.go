// Copyright 2018 The ZikiChomgo Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package entry

var names = [...]string{"CoreAudio Audio Queue Services", "CoreAudio AUHAL", "CoreAudio RemoteIO"}

func Names() []string {
	res := make([]string, len(names))
	copy(res, names)
	return res
}
