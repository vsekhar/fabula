variable "name" {
    type = string
}

variable "container_image" {
    type = object({
        project = string
        name = string
        digest = string
        image_url = string
    })
    description = "Container registry image data resource (e.g. 'google_container_registry_image')."
}

variable "args" {
    type = list(string)
    default = []
}

variable "host_to_container_ports" {
    type = map(string)
    default = {}
}

variable "network" {
    type = string
    default = null
}

variable "subnetwork" {
    type = string
    default = null
}

variable "public_ip" {
    type = string
    description = "Public IP to assign to the instance: '0.0.0.0' for ephemeral; '' for none (the default)"
    default = ""
}

variable "service_account" {
    type = string
    default = null
}

variable "preemptible" {
    type = bool
    default = false
}

variable "machine_type" {
    type = string
}

variable "tags" {
    type = list(string)
    default = []
}
