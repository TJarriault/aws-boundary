resource "aws_db_instance" "boundary-demo" {
  allocated_storage   = 20
  storage_type        = "gp2"
  engine              = "postgres"
  engine_version      = "13.4"
  instance_class      = "db.m5.large"
  name                = "devoxxdemo"
  identifier          = "devoxxdemo"
  username            = "boundary"
  password            = "boundarydemo"
  skip_final_snapshot = true

  vpc_security_group_ids = [aws_security_group.db-demo.id]
  db_subnet_group_name   = aws_db_subnet_group.boundary-demo.name
  publicly_accessible    = true

  tags = {
    Name = "${var.tag}-devoxx-demo-db"
  }
}

resource "aws_security_group" "db-demo" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.tag}-devoxx-demo-db-${random_pet.test.id}"
  }
}

resource "aws_security_group_rule" "allow_controller_sg-demo" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  security_group_id        = aws_security_group.db-demo.id
  source_security_group_id = aws_security_group.controller.id
}

resource "aws_security_group_rule" "allow_any_ingress-demo" {
  type              = "ingress"
  from_port         = 5432
  to_port           = 5432
  protocol          = "tcp"
  security_group_id = aws_security_group.db-demo.id
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_db_subnet_group" "boundary-demo" {
  name       = "boundary-devoxx-demo"
  subnet_ids = aws_subnet.public.*.id

  tags = {
    Name = "${var.tag}-devoxx-demo-db-${random_pet.test.id}"
  }
}
