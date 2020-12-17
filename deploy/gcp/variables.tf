variable "project_id" {
    type = string
}

variable "fabula_region" {
    type = string
    description = "The region to which Fabula should be deployed."
    validation {
        // keep in sync with ../regions
        condition = contains(["US", "EU", "JP", "SIN", "AUS", "IN"], var.fabula_region)
        error_message = "Must be one of: US, EU, JP, SIN, AUS, IN."
    }
}
