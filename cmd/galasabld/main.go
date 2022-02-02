//
// Licensed Materials - Property of IBM
//
// (c) Copyright IBM Corp. 2021.
//

package main

import (
	"os"
	"fmt"

	"galasa.dev/buildUtilities/pkg/cmd"
)

func main() {

	fmt.Println(os.Args)

	cmd.Execute()

}
