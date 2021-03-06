/*
*  Copyright (c) WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-resty/resty"
	"github.com/spf13/cobra"
	"github.com/wso2/product-apim-tooling/import-export-cli/utils"

	"github.com/renstrom/dedent"
	"net/http"
	"path/filepath"
)

var flagExportAPIName string
var flagExportAPIVersion string
var flagExportEnvironment string
var exportAPICmdUsername string
var exportAPICmdPassword string
var flagExportAPICmdToken string

// ExportAPI command related usage info
const exportAPICmdLiteral = "export-api"
const exportAPICmdShortDesc = "Export API"

var exportAPICmdLongDesc = "Export an API from an environment"

var exportAPICmdExamples = dedent.Dedent(`
		Examples:
		` + utils.ProjectName + ` ` + exportAPICmdLiteral + ` -n TwitterAPI -v 1.0.0 -e dev
		` + utils.ProjectName + ` ` + exportAPICmdLiteral + ` -n FacebookAPI -v 2.1.0 -e production
	`)

// ExportAPICmd represents the exportAPI command
var ExportAPICmd = &cobra.Command{
	Use: exportAPICmdLiteral + " (--name <name-of-the-api> --version <version-of-the-api> --environment " +
		"<environment-from-which-the-api-should-be-exported>)",
	Short: exportAPICmdShortDesc,
	Long:  exportAPICmdLongDesc + exportAPICmdExamples,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Logln(utils.LogPrefixInfo + exportAPICmdLiteral + " called")
		executeExportApiCmd(utils.MainConfigFilePath, utils.EnvKeysAllFilePath, utils.ExportDirectory)
	},
}

func executeExportApiCmd(mainConfigFilePath, envKeysAllFilePath, exportDirectory string) {
	if flagExportAPICmdToken != "" {
		// token provided with --token (-t) flag
		if exportAPICmdUsername != "" || exportAPICmdPassword != "" {
			// username and/or password provided with -u and/or -p flags
			// Error
			utils.HandleErrorAndExit("username/password provided with OAuth token.", nil)
		} else {
			// token only, proceed with token
			//publisherEndpoint := utils.GetPublisherEndpointOfEnv(flagExportEnvironment, utils.MainConfigFilePath)
			//ExportAPI(exportAPIName, exportAPIVersion, publisherEndpoint, exportAPICmdToken)
			fmt.Println("Token: " + flagExportAPICmdToken)
			publisherEndpoint := utils.GetPublisherEndpointOfEnv(flagExportEnvironment, mainConfigFilePath)
			accessToken := flagExportAPICmdToken
			resp := getExportApiResponse(flagExportAPIName, flagExportAPIVersion, publisherEndpoint, accessToken)

			// Print info on response
			utils.Logln(utils.LogPrefixInfo + "ExportAPI-ResponseStatus: " + resp.Status())

			if resp.StatusCode() == http.StatusOK {
				WriteToZip(flagExportAPIName, flagExportAPIVersion, flagExportEnvironment, exportDirectory, resp)

				// only to get the number of APIs exported
				numberOfAPIsExported, _, err := GetAPIList(flagExportAPIName, accessToken, publisherEndpoint)
				if err == nil {
					fmt.Println("Number of APIs exported:", numberOfAPIsExported)
				} else {
					utils.HandleErrorAndExit("Error getting list of APIs", err)
				}
			} else if resp.StatusCode() == http.StatusInternalServerError {
				// 500 Internal Server Error
				fmt.Println("Incorrect password")
			} else {
				// neither 200 nor 500
				fmt.Println("Error exporting API:", resp.Status())
			}
		}
	} else {
		// no token provided with --token (-t) flag
		// proceed with username and password
		accessToken, publisherEndpoint, preCommandErr := utils.ExecutePreCommand(flagExportEnvironment,
			exportAPICmdUsername, exportAPICmdPassword, mainConfigFilePath, envKeysAllFilePath)

		if preCommandErr == nil {
			resp := getExportApiResponse(flagExportAPIName, flagExportAPIVersion, publisherEndpoint, accessToken)

			// Print info on response
			utils.Logln(utils.LogPrefixInfo + "ExportAPI-ResponseStatus: " + resp.Status())

			if resp.StatusCode() == http.StatusOK {
				WriteToZip(flagExportAPIName, flagExportAPIVersion, flagExportEnvironment, exportDirectory, resp)

				// only to get the number of APIs exported
				numberOfAPIsExported, _, err := GetAPIList(flagExportAPIName, accessToken, publisherEndpoint)
				if err == nil {
					fmt.Println("Number of APIs exported:", numberOfAPIsExported)
				} else {
					utils.HandleErrorAndExit("Error getting list of APIs", err)
				}
			} else if resp.StatusCode() == http.StatusInternalServerError {
				// 500 Internal Server Error
				fmt.Println("Incorrect password")
			} else {
				// neither 200 nor 500
				fmt.Println("Error exporting API:", resp.Status())
			}
		} else {
			// error exporting API
			fmt.Println("Error exporting API:" + preCommandErr.Error())
		}
	}
}

// WriteToZip
// @param exportAPIName : Name of the API to be exported
// @param resp : Response returned from making the HTTP request (only pass a 200 OK)
// Exported API will be written to a zip file
func WriteToZip(exportAPIName, exportAPIVersion, exportEnvironment, exportDirectory string, resp *resty.Response) {
	// Write to file
	directory := filepath.Join(exportDirectory, exportEnvironment)
	// create directory if it doesn't exist
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0777)
		// permission 777 : Everyone can read, write, and execute
	}
	zipFilename := exportAPIName + "_" + exportAPIVersion + ".zip" // MyAPI_1.0.0.zip
	pFile := filepath.Join(directory, zipFilename)
	err := ioutil.WriteFile(pFile, resp.Body(), 0644)
	// permission 644 : Only the owner can read and write.. Everyone else can only read.
	if err != nil {
		utils.HandleErrorAndExit("Error creating zip archive", err)
	}
	fmt.Println("Succesfully exported API!")
}

// ExportAPI
// @param name : Name of the API to be exported
// @param version : Version of the API to be exported
// @param apimEndpoint : API Manager Endpoint for the environment
// @param accessToken : Access Token for the resource
// @return response Response in the form of *resty.Response
func getExportApiResponse(name string, version string, publisherEndpoint string, accessToken string) *resty.Response {
	// append '/' to the end if there isn't one already
	if string(publisherEndpoint[len(publisherEndpoint)-1]) != "/" {
		publisherEndpoint += "/"
	}

	publisherEndpoint += "export/apis?query="

	query := ""
	if name != "" {
		query += name
		// TODO: add version to the query after making sure carbon-apimgt backend supports it
	}

	url := publisherEndpoint + query
	utils.Logln(utils.LogPrefixInfo+"ExportAPI: URL:", url)
	headers := make(map[string]string)
	headers[utils.HeaderAuthorization] = utils.HeaderValueAuthPrefixBearer + " " + accessToken
	headers[utils.HeaderAccept] = utils.HeaderValueApplicationZip

	resp, err := utils.InvokeGETRequest(url, headers)

	if err != nil {
		utils.HandleErrorAndExit("Error exporting API: "+name, err)
	}

	return resp
}

// init using Cobra
func init() {
	RootCmd.AddCommand(ExportAPICmd)
	ExportAPICmd.Flags().StringVarP(&flagExportAPIName, "name", "n", "",
		"Name of the API to be exported")
	ExportAPICmd.Flags().StringVarP(&flagExportAPIVersion, "version", "v", "",
		"Version of the API to be exported")
	ExportAPICmd.Flags().StringVarP(&flagExportEnvironment, "environment", "e",
		utils.DefaultEnvironmentName, "Environment to which the API should be exported")
	ExportAPICmd.Flags().StringVarP(&flagExportAPICmdToken, "token", "t", "",
		"An OAuth2 token to be used instead of username and password")

	ExportAPICmd.Flags().StringVarP(&exportAPICmdUsername, "username", "u", "", "Username")
	ExportAPICmd.Flags().StringVarP(&exportAPICmdPassword, "password", "p", "", "Password")
}
