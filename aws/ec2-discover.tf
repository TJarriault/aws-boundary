variable "instances" {
  default = [
    "boundary-discovery-1-dev", 
    "boundary-discovery-2-dev", 
    "boundary-discovery-3-production", 
    "boundary-discovery-4-production"
  ]
}

variable "vm_tags" {
  default = [
    {"Name":"boundary-discovery-1-dev","service-type":"database", "application":"dev"},
    {"Name":"boundary-discovery-2-dev","service-type":"database", "application":"dev"},
    {"Name":"boundary-discovery-3-production","service-type":"database", "application":"production"},
    {"Name":"boundary-discovery-4-production","service-type":"database", "application":"prod"}
  ]
}

resource "aws_security_group" "boundary-ssh" {
vpc_id = aws_vpc.main.id
  name        = "boundary_allow_ssh"
  description = "Allow SSH inbound traffic"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_ssh"
  }
}

resource "aws_instance" "boundary-instance" {
  count                  = length(var.instances)
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.private.*.id[1]
  key_name               = aws_key_pair.boundary.key_name

  vpc_security_group_ids = [aws_security_group.boundary-ssh.id]
  tags                   = var.vm_tags[count.index]
}
