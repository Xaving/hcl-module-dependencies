// from the https://raw.githubusercontent.com/aamkye/bday/master/terraform/main.tf
provider "aws" {
  profile = var.aws-profile
  region  = "eu-west-1"
  default_tags {
    tags = {
      Env = "bday"
    }
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
  backend "s3" {
    profile = "bday"
    bucket  = "ak95-terraform-state"
    encrypt = true
    key     = "terraform.tfstate"
    region  = "eu-west-1"
    dynamodb_table = "terraform-state-lock-dynamo"
  }
}

module "vpc" {
  source             = "./vpc"
  availability_zones = var.availability_zones
  cidr               = var.cidr
  private_subnets    = var.private_subnets
  public_subnets     = var.public_subnets
}

module "sg" {
  source = "./sg"
  vpc_id = module.vpc.id
}

module "docdb" {
  source             = "./docdb"
  private_subnets    = module.vpc.private_subnets
  mongo_sg           = module.sg.mongo
  dbusername         = var.dbusername
  dbpassword         = var.dbpassword
  availability_zones = var.availability_zones
}

module "alb" {
  source            = "./alb"
  alb_sg            = module.sg.alb
  public_subnets    = module.vpc.public_subnets
  vpc_id            = module.vpc.id
  alb_tls_cert_arn  = var.alb_tls_cert_arn
  container_port    = var.container_port
  health_check_path = var.health_check_path
}

module "ecs" {
  source                = "./ecs"
  alb_tg_arn            = module.alb.tg_arn
  docdb_endpoint        = module.docdb.endpoint
  private_subnets       = module.vpc.private_subnets
  sg_ecs                = module.sg.ecs
  container_cpu         = var.container_cpu
  container_image       = var.container_image
  container_memory      = var.container_memory
  container_port        = var.container_port
  dbpassword            = var.dbpassword
  dbusername            = var.dbusername
  service_desired_count = var.service_desired_count
}

module "r53" {
  source          = "./r53"
  r53_zone_id     = var.r53_zone_id
  r53_record_name = var.r53_record_name
  alias_dns_name  = module.alb.dns_name
  alias_zone_id   = module.alb.zone_id
}
