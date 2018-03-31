/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"bufio"
	"os"
)

type inputEvent struct {
	done chan bool
}

// WaitForEnter waits until the user presses Enter
func (c *inputEvent) WaitForEnter() chan bool {
	go c.readFromCLI()
	return c.done
}

func (c *inputEvent) readFromCLI() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	c.done <- true
}
