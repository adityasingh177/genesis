/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package entity

import (
	"strconv"
	"testing"
)

func TestResources_GetMemory_Successful(t *testing.T) {
	var tests = []struct {
		res      Resources
		expected int64
	}{
		{res: Resources{
			Cpus:   "",
			Memory: "45",
		}, expected: int64(45)},
		{res: Resources{
			Cpus:   "",
			Memory: "1",
		}, expected: int64(1)},
		{res: Resources{
			Cpus:   "",
			Memory: "92233720368547",
		}, expected: int64(92233720368547)},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			num, _ := tt.res.GetMemory()

			if num != tt.expected {
				t.Error("return value of GetMemory does not match expected value")
			}
		})
	}
}

func TestResources_GetMemory_Unsuccessful(t *testing.T) {
	var tests = []struct {
		res Resources
	}{
		{res: Resources{
			Cpus:   "",
			Memory: "45.46",
		}},
		{res: Resources{
			Cpus:   "",
			Memory: "35273409857203948572039458720349857",
		}},
		{res: Resources{
			Cpus:   "",
			Memory: "s",
		}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := tt.res.GetMemory()

			if err == nil {
				t.Error("return error of GetMemory does not match expected error")
			}
		})
	}
}