/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	admin "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	"github.com/hyperledger/fabric/common/cauthdsl"
	fabricCommon "github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var instantiateCmd = &cobra.Command{
	Use:   "instantiate",
	Short: "Instantiate chaincode.",
	Long:  "Instantiates the chaincode",
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
		action, err := newInstantiateAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing instantiateAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running instantiateAction: %v", err)
		}
	},
}

func getInstantiateCmd() *cobra.Command {
	flags := instantiateCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	common.Config().InitChaincodePath(flags)
	common.Config().InitChaincodeVersion(flags)
	common.Config().InitArgs(flags)
	common.Config().InitChaincodePolicy(flags)
	return instantiateCmd
}

type instantiateAction struct {
	common.Action
}

func newInstantiateAction(flags *pflag.FlagSet) (*instantiateAction, error) {
	action := &instantiateAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, fmt.Errorf("a peer must be specified")
	}
	return action, err
}

func (action *instantiateAction) invoke() error {
	channelClient, err := action.ChannelClient()
	if err != nil {
		return fmt.Errorf("Error getting channel client: %v", err)
	}

	argBytes := []byte(common.Config().Args())
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshalling JSON arg string: %v", err)
	}

	eventHub, err := action.EventHub()
	if err != nil {
		return err
	}

	// InstantiateCC requires AdminUser privileges so setting user context with Admin User
	context := action.SetUserContext(action.OrgAdminUser(action.OrgID()))
	defer context.Restore()

	common.Config().Logger().Infof("Sending instantiate %s ...\n", common.Config().ChaincodeID())

	chaincodePolicy, err := action.newChaincodePolicy()
	if err != nil {
		return err
	}

	err = admin.SendInstantiateCC(
		channelClient, common.Config().ChaincodeID(), args.Args, common.Config().ChaincodePath(), common.Config().ChaincodeVersion(),
		chaincodePolicy, []apitxn.ProposalProcessor{action.Peer()}, eventHub)
	if err != nil {
		if strings.Contains(err.Error(), "chaincode exists "+common.Config().ChaincodeID()) {
			// Ignore
			common.Config().Logger().Infof("Chaincode %s already instantiated.", common.Config().ChaincodeID())
			fmt.Printf("...chaincode %s already instantiated.\n", common.Config().ChaincodeID())
			return nil
		}
		return fmt.Errorf("error instantiating chaincode: %v", err)
	}

	fmt.Printf("...successfuly instantiated chaincode %s on channel %s.\n", common.Config().ChaincodeID(), common.Config().ChannelID())

	return nil
}

func (action *instantiateAction) newChaincodePolicy() (*fabricCommon.SignaturePolicyEnvelope, error) {
	if common.Config().ChaincodePolicy() != "" {
		// Create a signature policy from the policy expression passed in
		return newChaincodePolicy(common.Config().ChaincodePolicy())
	}

	// Default policy is 'signed my any member' for all known orgs
	var mspIDs []string
	for _, orgID := range common.Config().OrgIDs() {
		mspID, err := common.Config().MspID(orgID)
		if err != nil {
			return nil, fmt.Errorf("Unable to get the MSP ID from org ID %s: %s", orgID, err)
		}
		mspIDs = append(mspIDs, mspID)
	}
	return cauthdsl.SignedByAnyMember(mspIDs), nil
}

func newChaincodePolicy(policyString string) (*fabricCommon.SignaturePolicyEnvelope, error) {
	ccPolicy, err := cauthdsl.FromString(policyString)
	if err != nil {
		return nil, fmt.Errorf("invalid chaincode policy [%s]: %s", policyString, err)
	}
	return ccPolicy, nil
}
