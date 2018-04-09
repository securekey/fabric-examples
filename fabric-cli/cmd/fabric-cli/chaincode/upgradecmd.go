/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	fabricCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade chaincode.",
	Long:  "Upgrades the chaincode",
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
		action, err := newUpgradeAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing upgradeAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running upgradeAction: %v", err)
		}
	},
}

func getUpgradeCmd() *cobra.Command {
	flags := upgradeCmd.Flags()
	cliconfig.InitPeerURL(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitChaincodePath(flags)
	cliconfig.InitChaincodeVersion(flags)
	cliconfig.InitArgs(flags)
	cliconfig.InitChaincodePolicy(flags)
	cliconfig.InitTimeout(flags)
	return upgradeCmd
}

type upgradeAction struct {
	action.Action
}

func newUpgradeAction(flags *pflag.FlagSet) (*upgradeAction, error) {
	action := &upgradeAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, errors.Errorf("a peer must be specified")
	}
	return action, err
}

func (a *upgradeAction) invoke() error {
	argBytes := []byte(cliconfig.Config().Args())
	args := &action.ArgStruct{}

	if err := json.Unmarshal(argBytes, args); err != nil {
		return errors.Errorf("Error unmarshalling JSON arg string: %v", err)
	}

	resMgmtClient, err := a.ResourceMgmtClient()
	if err != nil {
		return err
	}

	cliconfig.Config().Logger().Infof("Sending upgrade %s ...\n", cliconfig.Config().ChaincodeID())

	chaincodePolicy, err := a.newChaincodePolicy()
	if err != nil {
		return err
	}

	req := resmgmt.UpgradeCCRequest{
		Name:    cliconfig.Config().ChaincodeID(),
		Path:    cliconfig.Config().ChaincodePath(),
		Version: cliconfig.Config().ChaincodeVersion(),
		Args:    utils.AsBytes(args.Args),
		Policy:  chaincodePolicy,
	}

	_, err = resMgmtClient.UpgradeCC(cliconfig.Config().ChannelID(), req, resmgmt.WithTargets(a.Peers()...))
	if err != nil {
		if strings.Contains(err.Error(), "chaincode exists "+cliconfig.Config().ChaincodeID()) {
			// Ignore
			cliconfig.Config().Logger().Infof("Chaincode %s already instantiated.", cliconfig.Config().ChaincodeID())
			fmt.Printf("...chaincode %s already instantiated.\n", cliconfig.Config().ChaincodeID())
			return nil
		}
		return errors.Errorf("error instantiating chaincode: %v", err)
	}

	fmt.Printf("...successfuly upgraded chaincode %s on channel %s.\n", cliconfig.Config().ChaincodeID(), cliconfig.Config().ChannelID())

	return nil
}

func (a *upgradeAction) newChaincodePolicy() (*fabricCommon.SignaturePolicyEnvelope, error) {
	if cliconfig.Config().ChaincodePolicy() != "" {
		// Create a signature policy from the policy expression passed in
		return newChaincodePolicy(cliconfig.Config().ChaincodePolicy())
	}

	// Default policy is 'signed my any member' for all known orgs
	var mspIDs []string
	for _, orgID := range cliconfig.Config().OrgIDs() {
		mspID, err := a.EndpointConfig().MSPID(orgID)
		if err != nil {
			return nil, errors.Errorf("Unable to get the MSP ID from org ID %s: %s", orgID, err)
		}
		mspIDs = append(mspIDs, mspID)
	}
	return cauthdsl.SignedByAnyMember(mspIDs), nil
}
