// Copyright (c) 2025 Tethys Plex
//
// This file is part of REXI-hub v2.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package common

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"os"
	"runtime/pprof"
	"time"
)

// Monitor định kỳ giám sát mức sử dụng CPU, vượt ngưỡng thì xuất tệp pprof
func Monitor() {
	for {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			panic(err)
		}
		if percent[0] > 80 {
			fmt.Println("cpu usage too high")
			// write pprof file
			if _, err := os.Stat("./pprof"); os.IsNotExist(err) {
				err := os.Mkdir("./pprof", os.ModePerm)
				if err != nil {
					SysLog("Failed to create pprof folder " + err.Error())
					continue
				}
			}
			f, err := os.Create("./pprof/" + fmt.Sprintf("cpu-%s.pprof", time.Now().Format("20060102150405")))
			if err != nil {
				SysLog("Failed to create pprof file " + err.Error())
				continue
			}
			err = pprof.StartCPUProfile(f)
			if err != nil {
				SysLog("Failed to start pprof " + err.Error())
				continue
			}
			time.Sleep(10 * time.Second) // profile for 30 seconds
			pprof.StopCPUProfile()
			f.Close()
		}
		time.Sleep(30 * time.Second)
	}
}
