/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install chaincode.",
	Long:  "Install chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if common.Config().ChaincodePath() == "" {
			fmt.Printf("\nMust specify the path of the chaincode\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInstallAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing installAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running installAction: %v", err)
		}
	},
}

func getInstallCmd() *cobra.Command {
	flags := installCmd.Flags()
	common.Config().InitPeerURL(flags, "", "The URL of the peer on which to install the chaincode, e.g. localhost:7051")
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	common.Config().InitChaincodePath(flags)
	common.Config().InitChaincodeVersion(flags)
	return installCmd
}

type installAction struct {
	common.Action
}

func newInstallAction(flags *pflag.FlagSet) (*installAction, error) {
	action := &installAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *installAction) invoke() error {
	for orgID, peers := range action.PeersByOrg() {
		fmt.Printf("Installing chaincode %s on org[%s] peers:\n", common.Config().ChaincodeID(), orgID)
		for _, peer := range peers {
			fmt.Printf("-- %s\n", peer.URL())
		}
		err := action.installChaincode(orgID, peers)
		if err != nil {
			return err
		}
	}

	return nil
}

func (action *installAction) installChaincode(orgID string, targets []apifabclient.Peer) error {
	context := action.SetUserContext(action.OrgAdminUser(orgID))
	defer context.Restore()

	var errors []error

	transactionProposalResponse, _, err := action.Client().InstallChaincode(common.Config().ChaincodeID(), common.Config().ChaincodePath(), common.Config().ChaincodeVersion(), nil, targets)
	if err != nil {
		return fmt.Errorf("InstallChaincode returned error: %v", err)
	}

	ccIDVersion := common.Config().ChaincodeID() + "." + common.Config().ChaincodeVersion()

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			if strings.Contains(v.Err.Error(), ccIDVersion+" exists") {
				// Ignore
				fmt.Printf("Chaincode %s already installed on peer: %s.\n", ccIDVersion, v.Endorser)
			} else {
				errors = append(errors, fmt.Errorf("installCC returned error from peer %s: %v", v.Endorser, v.Err))
			}
		} else {
			fmt.Printf("...successfuly installed chaincode %s on peer %s.\n", ccIDVersion, v.Endorser)
		}
	}

	if len(errors) > 0 {
		common.Config().Logger().Warningf("Errors returned from InstallCC: %v\n", errors)
		return errors[0]
	}

	return nil
}
