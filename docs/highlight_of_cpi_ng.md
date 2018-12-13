# Highlights of New softlayer CPI

## Refactored the project structure and code

Refactored the structure and eliminate redundant code. The core structure is designed in three layers:
- The first layer is Actions layer: includes CPI actions to response to director and call each services methods
- The second layer is these services that provide interfaces to actions as abstraction of SL resources operations. Four services.
- The last layer is Softlayer client layer which does SL request and is called by Services.

Every layer provide function to upper as interface. It specifies a set of operations that an object can provide. We can use these clean interfaces to generate test-dependent code.

## Moved to softlayer-go official library

Adopted official softlayer-go library to fix this issue.
In addition to the maintain reason, it also has other useful features.
softlayer-go is official maintained SoftLayer API Client library written in golang.
It has the following features:
- Firstly, its code is pre-generated using SoftLayer API metadata endpoint as input, that ensures 100% coverage of the API.
- Secondly, It supports calling the corresponding methods to set masks, filters for each API request. We no longer need write code to generate request query string. We only need to call these methods that softlayer-go maintained.
- All non-slice method parameters in softlayer-go are passed as pointers. This is to allow for optional values to be omitted by passing nil.
- Softlayer-go suport a complete network timeout / rate-limits error retry mechanism. You can set the client to retry the api requests if there is a timeout error.
For above reasons, we used it to replace the original.

## Switched to HTTP Registry to store VM user data

The bosh-registry is a job of bosh-release. So for support this function, it must set  registry job when deploy the director.

The Agent is initially driven by the CPI through the bosh-registry, and then by the Director through NATS-based messaging. The registry provides Director-side metadata to the Agent. When running ‘create_vm’, CPI firstly stores registry endpoint(director) information to softlayer metadata server. Then CPI will update VM settings to registry db. When finished setup, agent on VM will run and read settings for registry endpoint. This method replaces the way that using ssh access modify the settings file.

## Used counterfeiter to generate fake code

This point is about the improvement of testing method.
Firstly, based on the three-tier structure of the project, it can easy to use counterfeiter to generate fake dependent code. Because the counterfeiter, the fake code generating tool, use interfaces to generate fake implementations.
Fake Client code provide positive and negative result to SL services.
And Fake SL services could provide resources for CPI actions.

## Adopted 3-tiers retry strategy to improve reliability

The point is about 3-tiers retry strategy that we applied to CPI.
The 3-tiers retry strategy includes three functional design：
- The first tie is in softlayer-go.softlayer-go library covers API request timeout and network errors.For example, it retries current request when dialing to SL DNS server timeout.
- The second tie is in softlayer services. As shown in the list, SL services would call these methods to wait Softlayer API return the expected result with timeout threshold. When it is timeout, services will return a timeout error.
- The third tie of retry is in actions. Actions return error containing retry option. Director would retry action request if CPI responds this option ‘retry: true’. For example, when wait VM ready timeout, action ‘create_vm’ will return can_retry error to the director and the director will call it again.

These measures can significantly improve the reliability of CPI.

## Implemented more useful functionalities

- has_disk: Checks if disk still exists
- get_disks: Checks if the VM has required disks attached
- set_disk_metadata: Sets disk’s metadata to make it easier for operators to categorize disks when looking at the IaaS management console(Volumes’ notes)
- snapshot_disk: Takes a snapshot of the disk
- delete_snapshot: Deletes the disk snapshot
- info: Returns information about the CPI to help the Director to make decisions on which CPI to call for certain operations in a multi CPI scenario

## Simplified cloud properties definition

Docs: [cloud_properties_names_in_cpi_ng.md](cloud_properties_names_in_cpi_ng.md)

## Added thread id to easily distinguish all log messages for one CPI call

Docs: [distinguish_logs_by_thread_ID_in_cpi_ng.md](distinguish_logs_by_thread_ID_in_cpi_ng.md)

## Supports networks split in network definition

Docs: [bosh_multi_homed_vms_example.md](bosh_multi_homed_vms_example.md)
