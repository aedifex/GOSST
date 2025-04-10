variable "resource_group_location" {
  # default     = "westus2"
  default     = "eastus2"
  description = "Location of the resource group."
}

variable "prefix" {
  type        = string
  default     = "win-vm-iis"
  description = "Prefix of the resource name"
}