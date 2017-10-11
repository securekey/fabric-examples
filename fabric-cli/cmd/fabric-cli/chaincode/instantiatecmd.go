/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	admin "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	fabricCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	cauthdslparser "github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var instantiateCmd = &cobra.Command{
	Use:   "instantiate",
	Short: "Instantiate chaincode.",
	Long:  "Instantiates the chaincode",
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
		action, err := newInstantiateAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing instantiateAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running instantiateAction: %v", err)
		}
	},
}

func getInstantiateCmd() *cobra.Command {
	flags := instantiateCmd.Flags()
	cliconfig.InitPeerURL(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitChaincodePath(flags)
	cliconfig.InitChaincodeVersion(flags)
	cliconfig.InitArgs(flags)
	cliconfig.InitChaincodePolicy(flags)
	return instantiateCmd
}

type instantiateAction struct {
	common.Action
}

func newInstantiateAction(flags *pflag.FlagSet) (*instantiateAction, error) {
	action := &instantiateAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, errors.Errorf("a peer must be specified")
	}
	return action, err
}

func (action *instantiateAction) invoke() error {
	argBytes := []byte(cliconfig.Config().Args())
	args := &common.ArgStruct{}

	if err := json.Unmarshal(argBytes, args); err != nil {
		return errors.Errorf("Error unmarshalling JSON arg string: %v", err)
	}

	channelClient, err := action.AdminChannelClient()
	if err != nil {
		return errors.Errorf("Error getting admin channel client: %v", err)
	}

	cliconfig.Config().Logger().Infof("Sending instantiate %s ...\n", cliconfig.Config().ChaincodeID())

	chaincodePolicy, err := action.newChaincodePolicy()
	if err != nil {
		return err
	}

	eventHub, err := action.EventHub()
	if err != nil {
		return err
	}

	err = admin.SendInstantiateCC(
		channelClient, cliconfig.Config().ChaincodeID(), asBytes(args.Args), cliconfig.Config().ChaincodePath(), cliconfig.Config().ChaincodeVersion(),
		chaincodePolicy, action.ProposalProcessors(), eventHub)
	if err != nil {
		if strings.Contains(err.Error(), "chaincode exists "+cliconfig.Config().ChaincodeID()) {
			// Ignore
			cliconfig.Config().Logger().Infof("Chaincode %s already instantiated.", cliconfig.Config().ChaincodeID())
			fmt.Printf("...chaincode %s already instantiated.\n", cliconfig.Config().ChaincodeID())
			return nil
		}
		return errors.Errorf("error instantiating chaincode: %v", err)
	}

	fmt.Printf("...successfuly instantiated chaincode %s on channel %s.\n", cliconfig.Config().ChaincodeID(), cliconfig.Config().ChannelID())

	return nil
}

func (action *instantiateAction) newChaincodePolicy() (*fabricCommon.SignaturePolicyEnvelope, error) {
	if cliconfig.Config().ChaincodePolicy() != "" {
		// Create a signature policy from the policy expression passed in
		return newChaincodePolicy(cliconfig.Config().ChaincodePolicy())
	}

	// Default policy is 'signed my any member' for all known orgs
	var mspIDs []string
	for _, orgID := range cliconfig.Config().OrgIDs() {
		mspID, err := cliconfig.Config().MspID(orgID)
		if err != nil {
			return nil, errors.Errorf("Unable to get the MSP ID from org ID %s: %s", orgID, err)
		}
		mspIDs = append(mspIDs, mspID)
	}
	return cauthdsl.SignedByAnyMember(mspIDs), nil
}

func newChaincodePolicy(policyString string) (*fabricCommon.SignaturePolicyEnvelope, error) {
	ccPolicy, err := cauthdslparser.FromString(policyString)
	if err != nil {
		return nil, errors.Errorf("invalid chaincode policy [%s]: %s", policyString, err)
	}
	return ccPolicy, nil
}
