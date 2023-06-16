package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Terrafrom struct {
	tk *tfexec.Terraform
}

var Tf = &Terrafrom{}

func init() {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.0")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("error get the current dir %s", err)
	}

	workingDir := path.Join(dir, "infrastrure/terraform")
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	fmt.Printf("Terraform initialized %v", &state)

	Tf.tk = tf
}

func (t *Terrafrom) plan() {
	tf := t.tk
	ctx := context.Background()

	options := []tfexec.PlanOption{
		tfexec.VarFile("variables.tfvars"),
	}

	state, err := tf.Plan(ctx, options...)
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	log.Printf("Terraform plan state: %v", state)
}

func (t *Terrafrom) apply() {
	tf := t.tk
	ctx := context.Background()

	options := []tfexec.ApplyOption{
		tfexec.VarFile("variables.tfvars"),
	}

	err := tf.Apply(ctx, options...)
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	result := <-ctx.Done()

	log.Printf("Terraform applied! %v", result)
}
