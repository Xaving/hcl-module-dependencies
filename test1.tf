resource "not_to_be_treated" "my_data" {
  attr1							= "test1" 
  attr2							= "test2" 
}

module "module1" {
  source = "git::ssh://git@github.com/jareware/terraform-utils.git//aws_ec2_ebs_docker_host?ref=v11.0"

  hostname             = "my-host"
  ssh_private_key_path = "~/.ssh/id_rsa" 
  ssh_public_key_path  = "~/.ssh/id_rsa.pub"
  data_volume_id       = "test" 
}

output "not_to_be_treated_2" {
  description = "test"
  value       = "abc"
}
