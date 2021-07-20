# gcp-vm-cleaner

A small application that will cleanup tagged VMs in a GCP account.

This is designed to clean up node-e2e test environments that have been
forgotten, to avoid maintaining out of tree patches, this means that it is
configured based on a vm-name prefix, rather than a label.

### Configuration

* `PROJECT`: The GCP project the backend is watching.
* `GCP_SERVICE_ACCOUNT_KEY`: The GCP service account key the backend can use to
  authenticate with GCP. 
* `GCLOUD_POLL_INTERVAL`: The poll interval used to retrieve/delete vms.
  Defaults to 10 minutes. The value must be specified in Golang's [time duration
  format](https://golang.org/pkg/time/#ParseDuration).
* `VM_LIFETIME_DURATION`: The maximum lifetime of a VM before it is eligible for
  deletion. Defaults to 24 hours. The value must
  be specified in Golang's [time duration
  format](https://golang.org/pkg/time/#ParseDuration).
* `VM_NAME_PREFIX`: The vm `name` prefix that will be matched when filtering vms
  for removal. Defaults to `test-cos-beta-`.
