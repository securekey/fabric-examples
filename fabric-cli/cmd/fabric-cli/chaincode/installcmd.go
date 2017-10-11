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
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install chaincode.",
	Long:  "Install chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if cliconfig.Config().ChaincodePath() == "" {
			fmt.Printf("\nMust specify the path of the chaincode\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInstallAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing installAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running installAction: %v", err)
		}
	},
}

func getInstallCmd() *cobra.Command {
	flags := installCmd.Flags()
	cliconfig.Config().InitPeerURL(flags, "", "The URL of the peer on which to install the chaincode, e.g. grpcs://localhost:7051")
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitChaincodeID(flags)
	cliconfig.Config().InitChaincodePath(flags)
	cliconfig.Config().InitChaincodeVersion(flags)
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
		fmt.Printf("Installing chaincode %s on org[%s] peers:\n", cliconfig.Config().ChaincodeID(), orgID)
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
	var errors []error

	user, err := action.OrgAdminUser(orgID)
	if err != nil {
		return err
	}

	client, err := action.ClientForUser(orgID, user)
	if err != nil {
		return fmt.Errorf("error creating fabric client: %s", err)
	}

	transactionProposalResponse, _, err := client.InstallChaincode(cliconfig.Config().ChaincodeID(), cliconfig.Config().ChaincodePath(), cliconfig.Config().ChaincodeVersion(), nil, targets)
	if err != nil {
		return fmt.Errorf("InstallChaincode returned error: %v", err)
	}

	ccIDVersion := cliconfig.Config().ChaincodeID() + "." + cliconfig.Config().ChaincodeVersion()

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
		cliconfig.Config().Logger().Warningf("Errors returned from InstallCC: %v\n", errors)
		return errors[0]
	}

	return nil
}
