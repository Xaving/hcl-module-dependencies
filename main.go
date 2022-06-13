package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// FileMapping is a schema for Terraform file.
type FileMapping struct {
	ModuleMappings []ModuleMapping `hcl:"module,block"`
	Remain         hcl.Body        `hcl:",remain"`
}

// ModuleMapping is a schema for "module" block in Terraform file.
type ModuleMapping struct {
	Name   string   `hcl:"name,label"`
	Source string   `hcl:"source"`
	Remain hcl.Body `hcl:",remain"`
}

// LoadTerraformFiles loads terraform files from a given dir.
func main() {

	p := hclparse.NewParser()
	f, diags := p.ParseHCLFile("main.tf")
	if diags.HasErrors() {
		log.Fatalf("diags: %v", diags)
		return
	}

	fm := &FileMapping{}
	diags = gohcl.DecodeBody(f.Body, nil, fm)
	if diags.HasErrors() {
		fmt.Println("Error")
		return
	}

	for _, m := range fm.ModuleMappings {
		fmt.Println("Module", m.Name, m.Source)
	}
	// Print the structure
	fmt.Println("RESULT", fm)

}
