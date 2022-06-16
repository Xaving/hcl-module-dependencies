# Copyright (c) 2021 Oracle and/or its affiliates.
# All rights reserved. The Universal Permissive License (UPL), Version 1.0 as shown at http://oss.oracle.com/licenses/upl
# main.tf
#
# Purpose: Defines all the components related to environment


module "mainbizcomp" {

  source = "git::ssh://git@github.com/oracle-devrel/terraform-oci-cloudbricks-compartment.git?ref=v1.0.2"
  providers = {
    oci.home = oci.home
  }
  ######################################## PROVIDER SPECIFIC VARIABLES ######################################
  tenancy_ocid     = var.tenancy_ocid
  region           = var.region
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  ######################################## PROVIDER SPECIFIC VARIABLES ######################################
  ######################################## COMPARTMENT SPECIFIC VARIABLES ######################################
  is_root_parent          = var.mainbizcomp_is_root_parent
  root_compartment_ocid   = var.mainbizcomp_root_compartment_ocid
  compartment_name        = var.mainbizcomp_compartment_name
  compartment_description = var.mainbizcomp_compartment_description
  enable_delete           = var.mainbizcomp_enable_delete
  ######################################## COMPARTMENT SPECIFIC VARIABLES ######################################

}


module "hub01" {
  providers = {
    oci.home = oci.home
  }
  source     = "git::ssh://git@github.com/oracle-devrel/terraform-oci-cloudbricks-compartment.git?ref=v1.0.2"
  depends_on = [module.mainbizcomp]
  ######################################## PROVIDER SPECIFIC VARIABLES ######################################
  tenancy_ocid     = var.tenancy_ocid
  region           = var.region
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  ######################################## PROVIDER SPECIFIC VARIABLES ######################################
  ######################################## COMPARTMENT SPECIFIC VARIABLES ######################################
  parent_compartment_name = module.mainbizcomp.compartment.name
  compartment_name        = var.hub01_compartment_name
  compartment_description = var.hub01_compartment_description
  enable_delete           = var.hub01_enable_delete
  ######################################## COMPARTMENT SPECIFIC VARIABLES ######################################

}

