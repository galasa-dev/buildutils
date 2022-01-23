//
// Copyright contributors to the Galasa project 
//

package galasayaml

type Results struct {
	Tests []ResultsTest `yaml:"tests"`
}

type ResultsTest struct {
	Name          string
	Bundle        string
	Class         string
}

