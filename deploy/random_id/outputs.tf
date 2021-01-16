output "id" {
    // "^(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)$"
    value = replace(lower(random_id.rid.id), "_", "-")
}
