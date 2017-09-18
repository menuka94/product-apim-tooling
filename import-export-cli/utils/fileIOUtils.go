/*
*  Copyright (c) 2005-2017, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
*
*  WSO2 Inc. licenses this file to you under the Apache License,
*  Version 2.0 (the "License"); you may not use this file except
*  in compliance with the License.
*  You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied.  See the License for the
* specific language governing permissions and limitations
* under the License.
*/

package utils

import (
	"gopkg.in/yaml.v2"
	"errors"
	"io/ioutil"
	"fmt"
	"os"
)

func WriteConfigFile(c interface{}, envConfigFilePath string) {
	data, err := yaml.Marshal(&c)
	if err != nil {
		HandleErrorAndExit("Unable to create Env Configuration.", err)
	}

	err = ioutil.WriteFile(envConfigFilePath, data, 0644)
	if err != nil {
		HandleErrorAndExit("Unable to create Env Configuration.", err)
	}
}

// Read and return EnvKeysAll
func GetEnvKeysAllFromFile(filePath string) *EnvKeysAll {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading " + filePath)
		os.Create(filePath)
		data, err = ioutil.ReadFile(filePath)
	}

	var envKeysAll EnvKeysAll
	if err := envKeysAll.ParseEnvKeysFromFile(data); err != nil {
		fmt.Println(LogPrefixError + "parsing " + filePath)
		return nil
	}

	return &envKeysAll
}

// Read and return EnvEndpointsAll
func GetEnvEndpointsAllFromFile(filePath string) *EnvEndpointsAll {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		HandleErrorAndExit("File Not Found: "+filePath, err)
	}

	var envEndpointsAll EnvEndpointsAll
	if err := envEndpointsAll.ParseEnvEndpointsFromFile(data); err != nil {
		HandleErrorAndExit("Error parsing "+filePath, err)
	}

	return &envEndpointsAll
}

// Read and validate contents of env_endpoints_all.yaml
// will throw errors if the any of the lines is blank
func (envEndpointsAll *EnvEndpointsAll) ParseEnvEndpointsFromFile(data []byte) error {
	if err := yaml.Unmarshal(data, envEndpointsAll); err != nil {
		return err
	}
	for name, endpoints := range envEndpointsAll.Environments {
		if endpoints.APIManagerEndpoint == "" {
			return errors.New("Blank API Manager Endpoint for " + name)
		}
		if endpoints.RegistrationEndpoint == "" {
			return errors.New("Blank Registration Endpoint for " + name)
		}
		if endpoints.TokenEndpoint == "" {
			return errors.New("Blank Token Endpoint for " + name)
		}
	}
	return nil
}

// Read and validate contents of env_keys_all.yaml
// will throw errors if the any of the lines is blank
func (envKeysAll *EnvKeysAll) ParseEnvKeysFromFile(data []byte) error {
	if err := yaml.Unmarshal(data, envKeysAll); err != nil {
		return err
	}
	for name, keys := range envKeysAll.Environments {
		if keys.ClientID == "" {
			return errors.New("Blank ClientID for " + name)
		}
		if keys.ClientSecret == "" {
			return errors.New("Blank ClientSecret for " + name)
		}
	}
	return nil
}