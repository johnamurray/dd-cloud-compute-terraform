# ddcloud\_server\_nic

Represents an additional network adapters in an existing server.

## Example Usage

```
resource "ddcloud_server_nic" "server_nic_test" {
  server       = "${ddcloud_server.test_server.id}"
  private_ipv4 = "192.168.18.100"
  vlan         = "${ddcloud_vlan.test_vlan.id}"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) ID of the server to which the network adapter needs to be added
* `private_ipv4` - (Optional) IPv4 Address for the new network adapter. Exactly one of `private_ipv4` or `vlan` or both must be specified.
* `vlan` - (Optional) VLAN ID of the new network adapter. Exactly one of `private_ipv4` or `vlan` or both must be specified.

## Attribute Reference

There are currently no additional attributes for `ddcloud_server_nic`.
