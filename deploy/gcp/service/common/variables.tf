terraform {
    experiments = [module_variable_optional_attrs]
}

variable "name" {
    type = string
    description = "Service name"
}

variable "group" {
    type = object({
        name = string
        network = string
        subnetwork = string
    })
    description = "Service group to use."
}

variable "versions" {
    type = map(object({
        container_image = object({
            project = string
            name = string
            digest = string
            image_url = string
        })
        args = optional(list(string), [])
        env = optional(map(string), {})
        machine_type = string
        target_size = optional(object({
            fixed = number
            percent = number
        }), null)
        preemptible = optional(bool, false)
        service_account = optional(string, "")
        envoy_config = optional(object({
            service_name = string
            envoy_service_port = string
            backend_protocol = string
            backend_service_port = string
        }), null)
    }))
    description = "Versions to run."
}

variable "http_health_check_path" {
    type = string
    description = "Relative path that returns 200 OK when healthy (e.g. '/healthz')."
}

variable "http_health_check_port" {
    type = number
    default = 8080
    description = "Port on which to send HTTP health checks."
}

variable "min_replicas" {
    type = number
    default = 1
    description = "Minimum number of replicas per region."
}

variable "max_replicas" {
    type = number
    default = 10
}

variable "service_to_container_ports" {
    type = map(string)
    default = {}
}

variable "service_account" {
    type = string
    default = null
}
