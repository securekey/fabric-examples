/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"bufio"
	"fmt"
	"os"
)

type inputEvent struct {
	done chan bool
}

// WaitForEnter waits until the user presses Enter
func (c *inputEvent) WaitForEnter() {
	go c.readFromCLI()
	<-c.done
}

func (c *inputEvent) readFromCLI() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press <enter> to terminate")
	reader.ReadString('\n')
	c.done <- true
}
