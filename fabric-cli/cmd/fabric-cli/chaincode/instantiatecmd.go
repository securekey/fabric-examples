/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
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
	cliconfig.InitCollectionConfigFile(flags)
	cliconfig.InitTimeout(flags)
	return instantiateCmd
}

type instantiateAction struct {
	action.Action
}

func newInstantiateAction(flags *pflag.FlagSet) (*instantiateAction, error) {
	action := &instantiateAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, errors.Errorf("a peer must be specified")
	}
	return action, err
}

func (a *instantiateAction) invoke() error {
	s := cliconfig.Config().Args()
	argBytes := []byte(s)
	args := &action.ArgStruct{}

	if err := json.Unmarshal(argBytes, args); err != nil {
		return errors.Errorf("Error unmarshalling JSON arg string: %v", err)
	}

	resMgmtClient, err := a.ResourceMgmtClient()
	if err != nil {
		return err
	}

	cliconfig.Config().Logger().Infof("Sending instantiate %s ...\n", cliconfig.Config().ChaincodeID())

	chaincodePolicy, err := a.newChaincodePolicy()
	if err != nil {
		return err
	}

	// Private Data Collection Configuration
	// - see fixtures/config/pvtdatacollection.json for sample config file
	var collConfig []*pb.CollectionConfig
	collConfigFile := cliconfig.Config().CollectionConfigFile()
	if collConfigFile != "" {
		collConfig, err = getCollectionConfigFromFile(cliconfig.Config().CollectionConfigFile())
		if err != nil {
			return errors.Wrapf(err, "error getting private data collection configuration from file [%s]", cliconfig.Config().CollectionConfigFile())
		}
	}

	req := resmgmt.InstantiateCCRequest{
		Name:       cliconfig.Config().ChaincodeID(),
		Path:       cliconfig.Config().ChaincodePath(),
		Version:    cliconfig.Config().ChaincodeVersion(),
		Args:       utils.AsBytes(utils.NewContext(), args.Args),
		Policy:     chaincodePolicy,
		CollConfig: collConfig,
	}

	_, err = resMgmtClient.InstantiateCC(cliconfig.Config().ChannelID(), req, resmgmt.WithTargets(a.Peer()))
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

func (a *instantiateAction) newChaincodePolicy() (*common.SignaturePolicyEnvelope, error) {
	if cliconfig.Config().ChaincodePolicy() != "" {
		// Create a signature policy from the policy expression passed in
		return newChaincodePolicy(cliconfig.Config().ChaincodePolicy())
	}

	// Default policy is 'signed by any member' for all known orgs
	return cauthdsl.AcceptAllPolicy, nil
}

func newChaincodePolicy(policyString string) (*common.SignaturePolicyEnvelope, error) {
	ccPolicy, err := cauthdsl.FromString(policyString)
	if err != nil {
		return nil, errors.Errorf("invalid chaincode policy [%s]: %s", policyString, err)
	}
	return ccPolicy, nil
}

type collectionConfigJSON struct {
	Name              string `json:"name"`
	Policy            string `json:"policy"`
	RequiredPeerCount int32  `json:"requiredPeerCount"`
	MaxPeerCount      int32  `json:"maxPeerCount"`
}

func getCollectionConfigFromFile(ccFile string) ([]*pb.CollectionConfig, error) {
	fileBytes, err := ioutil.ReadFile(ccFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file [%s]", ccFile)
	}
	cconf := &[]collectionConfigJSON{}
	if err = json.Unmarshal(fileBytes, cconf); err != nil {
		return nil, errors.Wrapf(err, "error parsing collection configuration in file [%s]", ccFile)
	}
	return getCollectionConfig(*cconf)
}

func getCollectionConfig(cconf []collectionConfigJSON) ([]*pb.CollectionConfig, error) {
	ccarray := make([]*pb.CollectionConfig, 0, len(cconf))
	for _, cconfitem := range cconf {
		p, err := cauthdsl.FromString(cconfitem.Policy)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("invalid policy %s", cconfitem.Policy))
		}
		cpc := &pb.CollectionPolicyConfig{
			Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
				SignaturePolicy: p,
			},
		}
		cc := &pb.CollectionConfig{
			Payload: &pb.CollectionConfig_StaticCollectionConfig{
				StaticCollectionConfig: &pb.StaticCollectionConfig{
					Name:              cconfitem.Name,
					MemberOrgsPolicy:  cpc,
					RequiredPeerCount: cconfitem.RequiredPeerCount,
					MaximumPeerCount:  cconfitem.MaxPeerCount,
				},
			},
		}
		ccarray = append(ccarray, cc)
	}
	return ccarray, nil
}
