/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"
	"net/http"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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
	cliconfig.InitPeerURL(flags, "", "The URL of the peer on which to install the chaincode, e.g. localhost:7051")
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitChaincodePath(flags)
	cliconfig.InitChaincodeVersion(flags)
	cliconfig.InitGoPath(flags)
	return installCmd
}

type installAction struct {
	action.Action
}

func newInstallAction(flags *pflag.FlagSet) (*installAction, error) {
	action := &installAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *installAction) invoke() error {
	var lastErr error
	for orgID, peers := range action.PeersByOrg() {
		fmt.Printf("Installing chaincode %s on org[%s] peers:\n", cliconfig.Config().ChaincodeID(), orgID)
		for _, peer := range peers {
			fmt.Printf("-- %s\n", peer.URL())
		}
		err := action.installChaincode(orgID, peers)
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func (action *installAction) installChaincode(orgID string, targets []fab.Peer) error {

	resMgmtClient, err := action.ResourceMgmtClientForOrg(orgID)
	if err != nil {
		return err
	}

	ccPkg, err := gopackager.NewCCPackage(cliconfig.Config().ChaincodePath(), cliconfig.Config().GoPath())
	if err != nil {
		return err
	}
	req := resmgmt.InstallCCRequest{
		Name:    cliconfig.Config().ChaincodeID(),
		Path:    cliconfig.Config().ChaincodePath(),
		Version: cliconfig.Config().ChaincodeVersion(),
		Package: ccPkg,
	}
	responses, err := resMgmtClient.InstallCC(req, resmgmt.WithTargets(targets...))
	if err != nil {
		return errors.Errorf("InstallChaincode returned error: %v", err)
	}

	ccIDVersion := cliconfig.Config().ChaincodeID() + "." + cliconfig.Config().ChaincodeVersion()

	var errs []error
	for _, resp := range responses {
		if resp.Info == "already installed" {
			fmt.Printf("Chaincode %s already installed on peer: %s.\n", ccIDVersion, resp.Target)
		} else if resp.Status != http.StatusOK {
			errs = append(errs, errors.Errorf("installCC returned error from peer %s: %s", resp.Target, resp.Info))
		} else {
			fmt.Printf("...successfuly installed chaincode %s on peer %s.\n", ccIDVersion, resp.Target)
		}
	}

	if len(errs) > 0 {
		cliconfig.Config().Logger().Warnf("Errors returned from InstallCC: %v\n", errs)
		return errs[0]
	}

	return nil
}
