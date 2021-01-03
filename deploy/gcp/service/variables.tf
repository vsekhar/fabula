terraform {
    experiments = [module_variable_optional_attrs]
}

variable "name" {
    type = string
    description = "Service name"
}

variable "region" {
    type = string
    description = "Region to deploy service to."
}

variable "versions" {
    type = list(object({
        name = string
        instance_template = string
        target_size = optional(object({
            fixed = number
            percent = number
        }), null)
    }))
    description = "Versions to run."
}

variable "external_ip" {
    type = string
    default = ""
    description = "External IP address to bind service to. Empty string for no external IP (default), '0.0.0.0' for an ephemeral IP, or specify a static IP."
}

variable "internal_lb" {
    type = bool
    default = false
    description = "Allocate an internal load-balanced IP address to this service."
}

variable "service_account" {
    type = string
    default = ""
    description = "Email address of service account"
}

variable "region_health_checks" {
    type = list(string)
    description = "IDs of regional health checks (google_compute_region_health_check). At least one is required."
}

variable "global_health_check" {
    type = string
    description = "ID of global health check (google_compute_health_check). Required."
}

variable "min_replicas" {
    type = number
    default = 1
    description = "Minimum number of replicas per region (3*n replicas will be started)."
}

variable "max_replicas" {
    type = number
    default = 10
}

variable "pubsub_autoscale" {
    type = object({
        subscription = string
        single_instance_assignment = number
    })
    default = null
    description = "PubSub subscription ID and per-instance handle rate to scale with."
}
