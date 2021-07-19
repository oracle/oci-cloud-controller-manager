# List Of E2E Tests

This is manually curated list of tests running under the test suite. Please keep it updated.

---

## CCM Tests
Use FOCUS=\[ccm\]
* [Load Balancer](load_balancer.go)
    * Basic Load Balancer Test - Verify creation and modification.
    * ESIP - OCI LBaaS is not a pass through load balancer so ESIPP (External Source IP Presevation) is not possible, however, this test covers support for node-local routing (i.e. avoidance of a second hop).
    * End to End TLS - Verify Service annotation for specifying the TLS secret to install on the load balancer listeners and backendsets which have SSL enabled.
    * TLS for Backendset - Verify Service annotation for specifying the TLS secret to install only on the load balancer backendsets which have SSL enabled.
    * TLS for Listener - Verify Service annotation for specifying the TLS secret to install only on the load balancer listeners which have SSL enabled.
    * End to End TLS with different certificates - Verify Service annotation for specifying different TLS secret to install on the load balancer listeners and backendsets which have SSL enabled.
    * LB properties - Verify Service annotations for modifying health check config, connection idle timeout and LB shape.
    
* [Instances](instances.go)
    * Verify if instance exists
    * Get node addresses
    * Get provider Id of an instance
    * Get the type of an instance

* [Zones](zones.go)
    * Get non-empty zone by node name
    * Get non-empty zone by provider id

---

## Storage Tests
Use FOCUS=\[storage\]
* [Block Volume Creation](block_volume_creation.go)
    * Create a persistent volume claim for a block storage
    * Create a persistent volume claim for a block storage with Ext3 File System
    * Create a persistent volume claim for a block storage with no AD specified
    
* [Backup Restore](backup_restore.go)
    * Backup a volume and restore the created backup

* [CSI Volume Creation](csi_volume_creation.go)
    * Create PVC and POD for CSI
    * Create PVC with VolumeSize 1Gi but should use default 50Gi
    * Create PVC with VolumeSize 100Gi should use 100G
    * Static Provisioning CSI
    * CMEK, PV attachment, in-transit encryption with CSI
    * CMEK, iSCSI attachment, in-transit encryption with CSI
    
* [Flex Volume Driver](flexvolume_driver.go)
    * Mount a volume
