/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package chaincode

import (
	"fmt"
	"strings"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/securekey/fabric-examples/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install chaincode.",
	Long:  "Install chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.ChaincodeID == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if common.ChaincodePath == "" {
			fmt.Printf("\nMust specify the path of the chaincode\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInstallAction(cmd.Flags())
		if err != nil {
			common.Logger.Criticalf("Error while initializing installAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Logger.Criticalf("Error while running installAction: %v", err)
			return
		}
	},
}

// Cmd returns the install command
func getInstallCmd() *cobra.Command {
	flags := installCmd.Flags()
	flags.String(common.PeerFlag, "", "The URL of the peer on which to install the chaincode, e.g. localhost:7051")
	flags.StringVar(&common.ChannelID, common.ChannelIDFlag, common.ChannelID, "The channel ID")
	flags.StringVar(&common.ChaincodeID, common.ChaincodeIDFlag, "", "The chaincode ID")
	flags.StringVar(&common.ChaincodePath, common.ChaincodePathFlag, "", "The chaincode path")
	flags.StringVar(&common.ChaincodeVersion, common.ChaincodeVersionFlag, common.ChaincodeVersion, "The chaincode version")
	return installCmd
}

type installAction struct {
	common.ActionImpl
}

func newInstallAction(flags *pflag.FlagSet) (*installAction, error) {
	action := &installAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *installAction) invoke() error {
	fmt.Printf("Installing chaincode %s on peers:\n", common.ChaincodeID)
	for _, peer := range action.Peers() {
		fmt.Printf("-- %s\n", peer.GetURL())
	}

	err := installChaincode(action.Client(), action.Peers(), common.ChaincodeID, common.ChaincodePath, common.ChaincodeVersion)
	if err != nil {
		return err
	}

	fmt.Printf("...successfuly installed chaincode %s on peers:\n", common.ChaincodeID)
	for _, peer := range action.Peers() {
		fmt.Printf("-- %s\n", peer.GetURL())
	}

	return nil
}

func installChaincode(client fabricClient.Client, targets []fabricClient.Peer, chaincodeID string, chaincodePath string, chaincodeVersion string) error {
	var errors []error

	transactionProposalResponse, _, err := client.InstallChaincode(chaincodeID, chaincodePath, chaincodeVersion, nil, targets)
	if err != nil {
		return fmt.Errorf("InstallChaincode returned error: %v", err)
	}

	ccIDVersion := chaincodeID + "." + chaincodeVersion

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			if strings.Contains(v.Err.Error(), ccIDVersion+" exists") {
				// Ignore
				common.Logger.Infof("Chaincode %s already installed on peer: %s.\n", ccIDVersion, v.Endorser)
			} else {
				errors = append(errors, fmt.Errorf("installCC returned error from peer %s: %v", v.Endorser, v.Err))
			}
		} else {
			common.Logger.Infof("...successfuly installed chaincode %s on peer %s.\n", ccIDVersion, v.Endorser)
		}
	}

	if len(errors) > 0 {
		common.Logger.Warningf("Errors returned from InstallCC: %v\n", errors)
		return errors[0]
	}

	return nil
}
