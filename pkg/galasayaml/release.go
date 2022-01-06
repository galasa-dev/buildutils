//
// Licensed Materials - Property of IBM
//
// (c) Copyright IBM Corp. 2021.
//

package galasayaml

type Release struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string 
	Metadata struct {
		Name string 
	} 
	Release struct {
		Version string
	}
	Framework struct {
		Bundles []Bundle
	}
	Api struct {
		Bundles []Bundle
	}
	Managers struct {
		Bundles []Bundle
	}
	External struct {
		Bundles []Bundle
	}
}

type Bundle struct {
	Group         string
	Artifact      string
	Version       string
	Type          string
	Obr           bool
	Bom           bool
	Isolated      bool
	Mvp           bool
	Javadoc       bool
	Managerdoc    bool
	Codecoverage  bool
}

