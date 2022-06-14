package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
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

	////////////////
	// first part //
	////////////////

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

	// Print module name and source
	for _, m := range fm.ModuleMappings {
		fmt.Println("Module", m.Name, m.Source)
	}

	///////////////
	//Second Part//
	///////////////

	// Parse module source Address
	location := fm.ModuleMappings[0].Source
	parts := strings.Split(location, "/")
	owner := parts[3]
	repo := strings.Split(parts[4], ".")[0]
	// Assume that the module dir contains a main.tf and a variables.tf
	pathv := strings.Split(parts[6], "?")[0] + "/" + "variables.tf"
	//pathm := strings.Split(parts[6], "?")[0] + "/" + "main.tf"

	// Get the file from the repo main branch
	client := github.NewClient(nil)
	file, _, _, _ := client.Repositories.GetContents(context.Background(), owner, repo, pathv, nil)
	//fc, _ := *file.GetContent()
	//fe := file.Encoding
	//c, _ := base64.StdEncoding.DecodeString(*file.Content)
	fmt.Println(c)
}
