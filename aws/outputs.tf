output "boundary_lb" {
  value = aws_lb.controller.dns_name
}

output "target_ips" {
  value = aws_instance.target.*.private_ip
}

output "target_ips_group1" {
  value = aws_instance.targetgroup1.*.private_ip
}
output "target_ips_group11" {
  value = aws_instance.targetgroup11.*.private_ip
}


output "kms_recovery_key_id" {
  value = aws_kms_key.recovery.id
}
