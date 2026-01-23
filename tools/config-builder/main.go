/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0

network-builder - Fabric-X Network Builder

A CLI tool to build, configure, and manage Fabric-X networks.
Replaces the Ansible-based deployment system with a simpler, more maintainable Go application.
*/
package main

import (
	"os"

	"github.com/ethsign/fabric-x-network/tools/config-builder/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
