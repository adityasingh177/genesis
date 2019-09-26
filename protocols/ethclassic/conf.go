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

package ethclassic

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/protocols/ethereum"
	"github.com/whiteblock/genesis/protocols/helpers"
)

// EtcConf represents the settings for the etc build
type EtcConf struct {
	ethereum.BaseConfig
	Identity        string `json:"identity"`
	Name            string `json:"name"`
	DAOHFBlock      int64  `json:"daoHFBlock"`
	EIP155_160Block int64  `json:"eip155_160Block"`
	ECIP1017Block   int64  `json:"ecip1017Block"`
	ECIP1017Era     int64  `json:"ecip1017Era"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*EtcConf, error) {
	out := new(EtcConf)
	err := helpers.HandleBlockchainConfig(blockchain, data, out)
	if err != nil || data == nil {
		return out, err
	}

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, fmt.Errorf("incorrect type for initBalance given")
		}
	}

	return out, nil
}

//NewEtcConf creates the configuration for etc
func NewEtcConf(data map[string]interface{}) (*EtcConf, error) {
	out := new(EtcConf)
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}