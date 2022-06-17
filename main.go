package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
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

// TODO: deal with source of the form : ./module/newmod

// LoadTerraformFiles loads terraform files from a given dir.
func main() {

	///////////////////////////////
	// first part: parse tf file //
	//////////////////////////////

	p := hclparse.NewParser()
	f, diags := p.ParseHCLFile("test1.tf")
	if diags.HasErrors() {
		log.Fatalf("diags: %v", diags)
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
	// todo: parse branch,manage error,
	// manage source name like source = "github.com/slswt/modules//aws/services/api_gateway"
	// check if it's a repo with main or master

	for _, mod := range ms.ModuleMappings {

		fmt.Println("###########", parseModuleAddress(mod.Source), "\n")
		// Parse module source Address
		location := mod.Source
		fmt.Println(location)
		parts := strings.Split(location, "/")
		owner := parts[3]
		repo := strings.Split(parts[4], ".")[0]

		// parse the path
		var pathv string
		if strings.Count(location, "//") == 2 {
			// Assume that the module dir contains a main.tf and a variables.tf
			pathv = strings.Split(parts[6], "?")[0] + "/" + "variables.tf"
			//pathm := strings.Split(parts[6], "?")[0] + "/" + "main.tf"
		} else {
			pathv = "variables.tf"
		}
		fmt.Printf("Owner: %s\nRepo: %s\nPathv: %s\n", owner, repo, pathv)

		// Check the HEAD branch name
		hn := checkHeadName(owner, repo)

		// Fetch the file from the repo master branch
		data, err := fetcher(owner, repo, pathv, hn)
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
		fmt.Printf("\n\n\n##########################\n")
		fmt.Printf("Module\n")
		fmt.Printf("\tName: %s\n", m.Name)
		fmt.Printf("\tSource: %s\n", m.Source)
		fmt.Printf("\tVariable: \n")
		for _, v := range variables {
			fmt.Printf("\t\tVariable: %s\n", v)
		}
	}
}

func parseModuleAddress(source string) []string {
	// manage source of different form: ssh:
	//git@github.com/, ../module, ./module, github.com/,git::ssh://git@github.com/

	// trim prefix
	t := source
	for _, p := range []string{"git@github.com/", "github.com/", "git::ssh://git@github.com/"} {
		t = strings.TrimPrefix(t, p)
	}

	// split string with ?ref=
	s := strings.Split(t, "?ref=")

	// extract path if any
	var path string
	n := s[0]
	if strings.Contains(n, "//") {
		c := strings.Split(n, "//")
		path = c[1]
		n = c[0]
	}

	// extract owner, repo
	var owner string
	var repo string
	p := strings.Split(n, "/")
	owner = p[0]
	repo = strings.TrimSuffix(p[1], ".git")

	// check if a specific branch
	var branch string

	if len(s) != 1 {
		// A branch was provided
		branch = s[1]
	} else {
		// check if repo is main or master
		// by fetching a file variables.tf
		branch = "main"
		address := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/variable.tf",
			owner, repo, branch, path)

		_, err := http.Get(address)
		if err != nil {
			// change main to master
			branch = "master"
		}
	}

	return []string{owner, repo, branch, path}
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

func fetcher(owner, repo, path, headname string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		owner, repo, headname, path)

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

func checkHeadName(owner, repo string) string {
	// Check if branch is master or main or something else
	url := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	r, _ := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})

	// ... retrieves the branch pointed by HEAD
	ref, _ := r.Head()
	rs := strings.Split(ref.String(), "/")
	result := rs[len(rs)-1]

	// don't catch error return
	return result
}
