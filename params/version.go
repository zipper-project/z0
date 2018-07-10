// Copyright 2018 The zipper Authors
// This file is part of the z0 library.
//
// The z0 library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The z0 library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the z0 library. If not, see <http://www.gnu.org/licenses/>.

package params

import "fmt"

const (
	//ClientIdentifier Client identifier to advertise over the network
	ClientIdentifier = "z0"
)

const (
	// VersionMajor is Major version component of the current release
	VersionMajor = 0
	// VersionMinor is Minor version component of the current release
	VersionMinor = 1
	// VersionPatch is Patch version component of the current release
	VersionPatch = 0
	// VersionMeta is Version metadata to append to the version string
	VersionMeta = "unstable"
)

// Version holds the textual version string.
var Version = func() string {
	v := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}()
