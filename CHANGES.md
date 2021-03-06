# Changes

## v1.1.4

New features:

* `ddcloud_port_list` now also supports specifying a list of simple ports (#50).
* Build executable for 32-bit Windows (Go calls this windows-386).

## v1.1.3

New features:

* `ddcloud_address_list` now also supports specifying a list of simple IP addresses (#49).
* Enable specifying of custom URL for CloudControl API end point (#48).

Bug fixes:

* Server name and description can now be modified after deployment (#47).

Changes:

* The `ddcloud` provider will now automatically retry requests to the CloudControl API if they fail due to network errors (default value for `retry_count` is now 3).

Breaking changes:

* Environment variables that previously had a `DDCLOUD_` prefix will now have an `MCP_` prefix (required for [docker-machine-driver-ddcloud #4](https://github.com/DimensionDataResearch/docker-machine-driver-ddcloud/issues/4))

## v1.1.2

New features:

* The provider now logs its version information at startup (for diagnostic purposes).
* Add `ddcloud_vlan` data source (#45)

Bug fixes:

* #44: Order of child address or port lists is not preserved

Breaking changes:

* The `child_lists` property on `ddcloud_address_list` and `ddcloud_port_list` is now a Set instead of a list. If you have existing Terraform state for this property it will not be retained (sorry).

## v1.1.1

New features:

* Add support for automatically creating tag keys if they are not already defined.
* Fixed bug #42 - unable to create `ddcloud_firewall_rule` with source address list or destination address list.  
This was due to inconsistencies in the CloudControl API for firewall rules (create vs read returns different field structure). The CloudControl client for Go has been updated and the new version included in the Terraform provider.

## v1.1

New features:

* Implement address lists (`ddcloud_address_list` resource type).
* Implement port lists (`ddcloud_port_list` resource type).
* Add address list and port list support to `ddcloud_firewall_rule` resource type.
* Add support for additional server network adapters (`ddcloud_server_nic` resource type).
* Expose CPU speed and cores-per-socket on `ddcloud_server` resource type (`cpu_speed` and `cores_per_cpu` properties).

## v1.0.2

New features:

* Implemented server anti-affinity rules (`ddcloud_server_anti_affinity` resource type).

## v1.0.1

Fixes:

* Fix for incorrect behaviour when adding VIP pool members with a specific port ([#17](https://github.com/DimensionDataResearch/dd-cloud-compute-terraform/pull/17/))
* `ddcloud_virtual_listener`'s `ipv4` address property is now computable (and captured during create / read).
* If there are no available public IP addresses when creating a `ddcloud_virtual_listener` without explicitly specifying an IPv4 address for the listener (i.e. CloudControl will allocate an IPv4 address), the provider will now automatically allocate a public IP block (similar to the behaviour of `ddcloud_nat_rule`).

## v0.7

New features:

* Full support for load-balancer configuration:
  * `ddcloud_vip_node`
  * `ddcloud_vip_pool`
  * `ddcloud_vip_pool_member`
  * `ddcloud_virtual_listener`

Fixes:

* `ddcloud_nat` resource is now fully idempotent (previously, if the NAT rule went missing, then Terraform did not correctly detect this).

## v0.6

* Extended logging of CloudControl API requests and responses can now be enabled by setting the `MCP_EXTENDED_LOGGING` environment variable to any non-empty value.

## v.04

Fixes:

* Changes to `ddcloud_server.tag` and `ddcloud_server.disk` are now correctly detected.

New features:

* `ddcloud_server` can now be deployed from a customer image (use `customer_image_id` / `customer_image_name` instead of `os_image_id`, `os_image_name`).

Breaking changes:

* `ddcloud_server.osimage_id` and `ddcloud_server.osimage_name` have been renamed to `ddcloud_server.os_image_id` and `ddcloud_server.os_image_id_name` (this is to be consistent with `customer_image_id` and `customer_image_name`).

## v0.3

* `ddcloud_server` now has a `public_ipv4` attribute that is resolved by matching any NAT rule that targets the server's primary network adapter's private IPv4 address. If the NAT rule gets the private IPv4 address from the server (rather than the other way around) then this attribute will not be available until you run `terraform refresh`.
