package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// FileMapping is a schema for Terraform file.
type Modules struct {
	ModuleMappings []ModuleMapping `hcl:"module,block"`
	Remain         hcl.Body        `hcl:",remain"`
}

// ModuleMapping is a schema for "module" block in Terraform file.
type ModuleMapping struct {
	Name   string   `hcl:"name,label"`
	Source string   `hcl:"source"`
	Remain hcl.Body `hcl:",remain"`
}

//Structure for variable buffer file
type Variables struct {
	VariableMappings []VariableMapping `hcl:"variable,block"`
	Remain           hcl.Body          `hcl:",remain"`
}

// VariableMapping is a schema for "variable" block
type VariableMapping struct {
	Name   string   `hcl:"name,label"`
	Remain hcl.Body `hcl:",remain"`
}

// LoadTerraformFiles loads terraform files from a given dir.
func main() {

	///////////////////////////////
	// first part: parse tf file //
	//////////////////////////////

	p := hclparse.NewParser()
	f, diags := p.ParseHCLFile("main.tf")
	if diags.HasErrors() {
		log.Fatalf("diags: %v", diags)
		return
	}

	ms := &Modules{}
	diags = gohcl.DecodeBody(f.Body, nil, ms)
	if diags.HasErrors() {
		fmt.Println("Error")
		return
	}

	/////////////////////////////////////
	//Second Part: fetch file in github//
	/////////////////////////////////////
	// todo: parse branch,manage error

	// Parse module source Address
	location := ms.ModuleMappings[0].Source
	parts := strings.Split(location, "/")
	owner := parts[3]
	repo := strings.Split(parts[4], ".")[0]
	// Assume that the module dir contains a main.tf and a variables.tf
	pathv := strings.Split(parts[6], "?")[0] + "/" + "variables.tf"
	//pathm := strings.Split(parts[6], "?")[0] + "/" + "main.tf"
	//fmt.Printf("Owner: %s\nRepo: %s\nPathv: %s\n", owner, repo, pathv)

	// Fetch the file from the repo master branch
	data, err := fetcher(owner, repo, pathv)
	if err != nil {
		log.Fatal("Cannot request repo: %s", err)
	}

	/////////////////////////////////
	//Third Part: parse the buffer //
	////////////////////////////////
	variables, diags := parseVariable(data)
	if diags.HasErrors() {
		log.Fatalf("Error while parsing the buffer for variables")
	}

	/////////////////////////////////
	// Fourth Part: Format a result//
	/////////////////////////////////

	m := ms.ModuleMappings[0]
	fmt.Printf("Module\n")
	fmt.Printf("\tName: %s\n", m.Name)
	fmt.Printf("\tSource: %s\n", m.Source)
	fmt.Printf("\tVariable: \n")
	for _, v := range variables {
		fmt.Printf("\t\tVariable: %s\n", v)
	}
}

func parseVariable(input []byte) ([]string, hcl.Diagnostics) {
	// Create the parser from the buffer
	p := hclparse.NewParser()
	pi, diags := p.ParseHCL(input, "from_variable_file")
	if diags.HasErrors() {
		return nil, diags
	}

	// Get the variables from the parser
	vs := &Variables{}
	diags = gohcl.DecodeBody(pi.Body, nil, vs)
	if diags.HasErrors() {
		fmt.Println("Error")
		return nil, diags
	}

	// Return the list of variable name
	var variables []string
	for _, v := range vs.VariableMappings {
		variables = append(variables, v.Name)
	}
	return variables, diags
}

func fetcher(owner, repo, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/%s",
		owner, repo, path)

	r, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	var data []byte
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		data = append(data, s.Bytes()...)
		data = append(data, '\n')
	}

	return data, nil
}
