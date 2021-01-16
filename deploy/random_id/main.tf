resource "random_id" "rid" {
    byte_length = var.byte_length
    keepers = var.keepers
}
