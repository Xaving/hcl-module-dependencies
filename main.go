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

	///////////////////////////////
	// first part: parse tf file //
	//////////////////////////////

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

	/////////////////////////////////////
	//Second Part: fetch file in github//
	/////////////////////////////////////
	// todo: parse branch,manage error

	// Parse module source Address
	location := fm.ModuleMappings[0].Source
	parts := strings.Split(location, "/")
	owner := parts[3]
	repo := strings.Split(parts[4], ".")[0]
	// Assume that the module dir contains a main.tf and a variables.tf
	pathv := strings.Split(parts[6], "?")[0] + "/" + "variables.tf"
	//pathm := strings.Split(parts[6], "?")[0] + "/" + "main.tf"
	fmt.Printf("Owner: %s\nRepo: %s\nPathv: %s\n", owner, repo, pathv)

	// Fetchthe file from the repo master branch
	data, err := fectcher(owner, repo, pathv)
	if err != nil {
		log.Fatal("Cannot request repo: %s", err)
	}
	fmt.Println(string(data))
}

func fectcher(owner, repo, path string) ([]byte, error) {
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
