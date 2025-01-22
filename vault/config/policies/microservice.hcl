# Read access to secrets
path "secret/data/users/*/keys" {
  capabilities = ["read"]
}

# Deny all other paths
path "*" {
  capabilities = ["deny"]
}
